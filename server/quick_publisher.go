package server

import (
	"log"
	"net"
	"time"

	"gitee.com/sylixos/go-vsoa/protocol"
)

func (s *Server) qpublisher(servicePath string, timeOrTrigger any, pubs func(*protocol.Message, *protocol.Message)) {
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
				go s.qsendMessage(req, client.QAddr)
			}
		}
	}
}

// Normal channel Publish Message
func (s *Server) qsendMessage(req *protocol.Message, qAddr *net.UDPAddr) error {
	req.SetMessageType(protocol.TypePublish)

	req.SetReply(false)

	tmp, err := req.Encode(protocol.ChannelNormal)
	if err != nil {
		log.Panicln(err)
		return err
	}

	s.qln.WriteToUDP(tmp, qAddr)
	protocol.PutData(&tmp)

	return err
}
