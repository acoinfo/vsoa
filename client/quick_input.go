package client

import (
	"go-vsoa/protocol"
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
	// Do nothing in quick channel
}
