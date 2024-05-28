package server

import (
	"log"
	"net"
	"time"

	"gitee.com/sylixos/go-vsoa/protocol"
)

// publisher is a method of the VsoaServer struct that sends a publish message to all active clients subscribed to a specific service path at the specified time interval.
//
// Parameters:
// - servicePath: a string representing the service path to publish.
// - timeDriction: a time.Duration value representing the time interval between each publish message.
// - pubs: a function that takes two parameters: a pointer to a protocol.Message and a pointer to another protocol.Message. It is called to initialize the request message before publishing.
func (s *Server) publisher(servicePath string, timeOrTrigger any, pubs func(*protocol.Message, *protocol.Message)) {
	req := protocol.NewMessage()

	var ticker *time.Ticker
	isTrigger := false

	switch v := timeOrTrigger.(type) {
	case time.Duration:
		ticker = time.NewTicker(v)
		defer ticker.Stop()
	case chan struct{}:
		s.triggerChan[servicePath] = v
		isTrigger = true
	default:
		panic("Invalid type for timeOrTrigger")
	}

	for {
		if isTrigger {
			<-s.triggerChan[servicePath]
		} else {
			<-ticker.C
		}

		pubs(req, nil)

		for _, client := range s.clients {
			if client.Subscribes[servicePath] && client.Authed {
				//PUT URL into req otherwise client will not receive this publish
				req.URL = []byte(servicePath)
				s.sendMessage(req, client.Conn)
			}
		}
	}
}

// Normal channel Publish Message
func (s *Server) sendMessage(req *protocol.Message, conn net.Conn) error {
	req.SetMessageType(protocol.TypePublish)

	req.SetReply(false)

	tmp, err := req.Encode(protocol.ChannelNormal)
	if err != nil {
		log.Panicln(err)
		return err
	}

	if s.writeTimeout != 0 {
		conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))
	}

	_, err = conn.Write(tmp)
	protocol.PutData(&tmp)

	return err
}
