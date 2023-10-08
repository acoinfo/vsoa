package server

import (
	"go-vsoa/protocol"
	"log"
	"net"
	"time"
)

func (s *VsoaServer) qpublisher(servicePath string, timeDriction time.Duration, pubs func(*protocol.Message, *protocol.Message)) {
	req := protocol.NewMessage()
	pubs(req, nil)

	ticker := time.NewTicker(timeDriction)
	defer ticker.Stop()

	for range ticker.C {
		for _, client := range s.activeClients {
			if client.Subscribes[servicePath] {
				//PUT URL into req otherwise client will not receive this publish
				req.URL = []byte(servicePath)
				s.qsendMessage(req, client.QAddr)
			}
		}
	}
}

// Normal channel Publish Message
func (s *VsoaServer) qsendMessage(req *protocol.Message, qAddr *net.UDPAddr) error {
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
