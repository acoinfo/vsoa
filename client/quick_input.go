package client

import (
	"errors"
	"fmt"
	"go-vsoa/protocol"
	"io"
	"log"
	"net"
)

// quick channel will only receive server's publish in Quick channel
func (client *Client) qinput() {
	var err error

	for err == nil {
		res := protocol.NewMessage()

		err = res.Decode(client.qr)
		if err != nil {
			break
		}

		switch {
		case res.MessageType() == protocol.TypePublish:
			// TODO: father URL logic
			if act, ok := client.SubscribeList[string(res.URL)]; ok {
				act(res)
			} else {
				continue
			}
		default:
			continue
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
	//We need to cloes normal channel too.
	client.Conn.Close()
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
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}

	client.mutex.Unlock()

	if err != nil && !closing {
		log.Printf("VSOA: client quick channel protocol error: %v", err)
	}
}
