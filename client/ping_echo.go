package client

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/acoinfo/vsoa/protocol"
)

// pingLoop runs a loop to send ping messages to the server and handle the responses.
// If timeout counts lager than PingLost, it will try to reconnect the connection.
func (client *Client) pingLoop() {
	if client.option.PingInterval == 0 {
		return
	}

	if client.option.PingTurbo != 0 {
		go client.pingTurboLoop()
	}

	client.pingEchoLoop()
}

func (client *Client) pingEchoLoop() {
	IntervalTime := time.Duration(client.option.PingInterval) * time.Second
	ticker := time.NewTicker(IntervalTime)

	// If Server / Client close conn, kill the pingLoop
	for client.pingTimeoutCount < client.option.PingLost {
		<-ticker.C
		req := protocol.NewMessage()
		reply := protocol.NewMessage()
		Call := client.Go("", protocol.TypePingEcho, nil, req, reply, nil).Done
		pingTimerOut := time.NewTimer(time.Second * time.Duration(client.option.PingTimeout))
		select {
		case call := <-Call:
			if call.Error != nil {
				atomic.StoreInt32(&client.pingTimeoutCount, 0)
			}
			pingTimerOut.Stop()
		case <-pingTimerOut.C:
			atomic.AddInt32(&client.pingTimeoutCount, 1)
			// Will we delete pending? Maybe it's not needed.
			pingTimerOut.Stop()
		}
	}

	ticker.Stop()
	// This is for Reconnect net-work if Just no respond.(TCP ACK but server respond nothing)
	client.Close()
	client.reConnect("vsoa")
}

func (client *Client) pingTurboLoop() {
	IntervalTime := time.Duration(client.option.PingTurbo) * time.Millisecond
	ticker := time.NewTicker(IntervalTime)

	// If Server / Client close conn, kill the pingLoop
	for client.pingTimeoutCount < client.option.PingLost {
		<-ticker.C

		if len(client.pending) == 0 {
			continue
		}

		req := protocol.NewMessage()
		reply := protocol.NewMessage()
		Call := client.Go("", protocol.TypeNoop, nil, req, reply, nil).Done
		pingTimerOut := time.NewTimer(time.Millisecond * time.Duration(client.option.PingTurbo))
		select {
		case <-Call:
			pingTimerOut.Stop()
		case <-pingTimerOut.C:
			pingTimerOut.Stop()
		}
	}

	ticker.Stop()
}

// Client send PingEcho message
// Similar to RPC call
// Internal use.
func (client *Client) sendPingEcho(call *Call) {
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}

	if client.option.PingInterval < client.option.PingTimeout || client.option.PingLost == 0 {
		call.Error = ErrPingEcho
		client.mutex.Unlock()
		call.done()
		return
	}

	if client.pending == nil {
		client.pending = make(map[uint32]*Call)
	}

	seq := client.seq
	client.seq++
	client.pending[seq] = call

	req := protocol.NewMessage()
	req.SetMessageType(protocol.TypePingEcho)

	// This means nothing.
	req.SetMessageRpcMethod(protocol.RpcMethodGet)
	req.SetSeqNo(seq)

	req.URL = []byte(call.URL)
	req.Param = *call.Param
	req.Data = call.Data

	tmp, err := req.Encode(protocol.ChannelNormal)
	if err != nil {
		call = client.pending[seq]
		delete(client.pending, seq)
		call.Error = err
		client.mutex.Unlock()
		call.done()
		return
	}
	client.mutex.Unlock()

	_, err = client.Conn.Write(tmp)

	// TODO: add ping fault logic
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e.Err != nil {
				err = fmt.Errorf("net.OpError: %s", e.Err.Error())
			} else {
				err = errors.New("net.OpError")
			}

		}
		client.mutex.Lock()
		client.pingTimeoutCount++
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}
}

// Client send Noop message
// Similar to RPC call but without no reply
// Internal use.
func (client *Client) sendNoop(call *Call) {
	// no reply. call is complete.
	defer call.done()

	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		return
	}

	if client.option.PingInterval < client.option.PingTimeout || client.option.PingLost == 0 {
		call.Error = ErrPingEcho
		return
	}

	client.mutex.Lock()
	var tmpNoseq uint32
	if client.noseq == 0 {
		tmpNoseq = 1
		client.noseq = 2
	} else {
		tmpNoseq = client.noseq
		client.noseq = (client.noseq + 1) & 0xffff
	}
	tmpNoseq = tmpNoseq << 16
	client.mutex.Unlock()

	req := protocol.NewMessage()
	req.SetMessageType(protocol.TypeNoop)

	// This means nothing.
	req.SetMessageRpcMethod(protocol.RpcMethodGet)
	req.SetSeqNo(tmpNoseq)

	req.URL = []byte(call.URL)
	req.Param = *call.Param
	req.Data = call.Data

	tmp, err := req.Encode(protocol.ChannelNormal)
	if err != nil {
		call.Error = err
		return
	}

	_, err = client.Conn.Write(tmp)

	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e.Err != nil {
				err = fmt.Errorf("net.OpError: %s", e.Err.Error())
			} else {
				err = errors.New("net.OpError")
			}

		}
		client.mutex.Lock()
		client.pingTimeoutCount++
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
		}
		return
	}
}
