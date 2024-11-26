package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/go-sylixos/go-vsoa/protocol"
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
			if act, ok := client.SubscribeList[string(res.URL)]; ok {
				if act != nil {
					act(res)
				}
				client.regulatorUpdator(res)
				continue
			}
			if act, ok := client.SubscribeList[string(res.URL)+"/"]; ok {
				if act != nil {
					act(res)
				}
				client.regulatorUpdator(res)
				continue
			}
			if act, ok := client.SubscribeList[string(res.URL[:len(res.URL)-1])]; ok && strings.HasSuffix(string(res.URL), "/") {
				if act != nil {
					act(res)
				}
				client.regulatorUpdator(res)
				continue
			}
			routelen := 0
			savedAct := func(m *protocol.Message) {}
			for route, actor := range client.SubscribeList {
				// Find one handler and send
				if strings.HasSuffix(route, "/") &&
					strings.HasPrefix(string(res.URL), route) {
					if len(route) < routelen {
						continue
					} else {
						routelen = len(route)
						savedAct = actor
					}
				}
			}
			if savedAct != nil {
				savedAct(res)
			}
			client.regulatorUpdator(res)
			continue
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
			call.Error = nil
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

	err = client.Close()

	if e, ok := err.(*net.OpError); ok {
		if e.Addr != nil || e.Err != nil {
			err = fmt.Errorf("net.OpError: %s", e.Err.Error())
		} else {
			err = errors.New("net.OpError")
		}

	}
	if err == io.EOF {
		if client.closing {
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

	if err != nil && !client.closing && !client.shutdown {
		log.Printf("VSOA: client protocol error: %v", err)
	}
}
