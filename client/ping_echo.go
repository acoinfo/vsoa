package client

import (
	"errors"
	"fmt"
	"go-vsoa/protocol"
	"net"
	"time"
)

func (client *Client) pingLoop() {
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
		case <-Call:
			client.mutex.Lock()
			client.pingTimeoutCount = 0
			client.mutex.Unlock()
			pingTimerOut.Stop()
		case <-pingTimerOut.C:
			client.mutex.Lock()
			client.pingTimeoutCount++
			// Will we delete pending? Maybe it's not needed.
			client.mutex.Unlock()
			pingTimerOut.Stop()
		}
	}

	ticker.Stop()
	// This is for Reconnect net-work if Just no respond.(TCP ACK but server respond nothing)
	client.Close()
	client.reConnect("vsoa")
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
