package client

import (
	"strings"

	"gitee.com/sylixos/go-vsoa/protocol"
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
		default:
			continue
		}
	}
	// Do nothing in quick channel
}
