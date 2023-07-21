package client

import (
	"errors"
	"fmt"
	"go-vsoa/protocol"
	"io"
	"log"
	"net"
)

func (client *Client) input() {
	var err error

	for err == nil {
		res := protocol.NewMessage()

		err = res.Decode(client.r)
		if err != nil {
			break
		}

		// This is for normal channel publish
		if res.MessageType() == protocol.TypePublish {
			// TODO: father URL logic
			if act, ok := client.SubscribeList[string(res.URL)]; ok {
				act(res)
			} else {
				continue
			}
		}

		seq := res.SeqNo()
		var call *Call
		isServerMessage := (!res.IsReply())
		if !isServerMessage {
			client.mutex.Lock()
			call = client.pending[seq]
			delete(client.pending, seq)
			client.mutex.Unlock()
		}

		switch {
		case call == nil:
			if isServerMessage {
				if client.ServerMessageChan != nil {
					client.handleServerRequest(res)
				}
				continue
			}
		case res.StatusType() != protocol.StatusSuccess:
			// We've got an error response. Give this to the request
			call.Error = strErr(res.StatusTypeText())
			call.Reply = res
			call.done()
		case res.StatusType() == protocol.StatusPassword:
			// We've got Passwd error response. Shutdown client
			call.Error = strErr(res.StatusTypeText())
			call.Reply = res
			call.done()
			client.Close()
		case res.IsPingEcho():
			fallthrough
		default:
			call.Reply = res
			call.done()
		}
	}

	// Terminate pending calls.
	// This is used for Subscribe in VSOA
	if client.ServerMessageChan != nil {
		req := protocol.NewMessage()
		req.SetMessageType(protocol.TypePublish)
		req.SetStatusType(protocol.StatusNoResponding)
		client.handleServerRequest(req)
	}

	client.mutex.Lock()
	client.Conn.Close()
	//We need to cloes quick channel too.
	client.QConn.Close()
	client.shutdown = true
	closing := client.closing
	if e, ok := err.(*net.OpError); ok {
		if e.Addr != nil || e.Err != nil {
			err = fmt.Errorf("net.OpError: %s", e.Err.Error())
		} else {
			err = errors.New("net.OpError")
		}

	}
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			// If server aggressive close the client conn.
			// TODO: reConnection logic
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}

	client.mutex.Unlock()

	if err != nil && !closing {
		log.Printf("VSOA: client protocol error: %v", err)
	}
}
