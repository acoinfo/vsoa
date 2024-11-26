package client

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/go-sylixos/go-vsoa/protocol"
)

// Subscribe server URL;
// inject onPublish callback to the URL with father URL
func (client *Client) Subscribe(URL string, onPublish func(m *protocol.Message)) error {
	var err error = nil
	if !client.IsAuthed() {
		err = ErrUnAuthed
		return err
	}

	if client.SubscribeList == nil {
		client.SubscribeList = make(map[string]func(*protocol.Message))
	}

	req := protocol.NewMessage()
	// Subscribe call the server will return nothing in Param & Data
	_, err = client.Call(URL, protocol.TypeSubscribe, nil, req)

	if err != nil {
		return err
	} else {
		client.mutex.Lock()
		defer client.mutex.Unlock()
		if onPublish != nil {
			client.SubscribeList[URL] = onPublish
		} else {
			client.SubscribeList[URL] = defaultOnPublish
		}
	}
	return err
}

// UnSubscribe server URL;
// free callback to the URL with father URL
func (client *Client) UnSubscribe(URL string) error {
	var err error = nil
	if !client.IsAuthed() {
		err = ErrUnAuthed
		return err
	}

	err = client.UnSlot(URL)
	if err != nil {
		log.Println(err)
	}

	if client.SubscribeList == nil {
		// Already unSubscribe
		return nil
	}

	if _, ok := client.SubscribeList[URL]; ok {
		client.mutex.Lock()
		delete(client.SubscribeList, URL)
		client.mutex.Unlock()
	} else if strings.HasSuffix(URL, "/") {
		if client.SubscribeList[URL[:len(URL)-1]] != nil {
			client.mutex.Lock()
			delete(client.SubscribeList, URL[:len(URL)-1])
			client.mutex.Unlock()
		}
		// Already unSubscribe
		return nil
	}

	req := protocol.NewMessage()
	// Subscribe call the server will return nothing in Param & Data
	_, err = client.Call(URL, protocol.TypeUnsubscribe, nil, req)

	if err != nil {
		return err
	}
	return err
}

func defaultOnPublish(m *protocol.Message) {
	log.Println("URL:", m.URL, "Param:", (m.Param), "Data:", (m.Data))
}

// Client send Sub/UnSub message
// Similar to RPC call
// Internal use. User should call Subscribe
func (client *Client) sendSubscribe(call *Call, isSubscribe bool) {
	// If it's Subscribe call Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
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
	if isSubscribe {
		req.SetMessageType(protocol.TypeSubscribe)
	} else {
		req.SetMessageType(protocol.TypeUnsubscribe)
	}

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

	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e.Err != nil {
				err = fmt.Errorf("net.OpError: %s", e.Err.Error())
			} else {
				err = errors.New("net.OpError")
			}

		}
		client.mutex.Lock()
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
