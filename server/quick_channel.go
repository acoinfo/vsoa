package server

import (
	"bytes"
	"errors"
	"go-vsoa/protocol"
	"io"
	"log"
	"net"
	"runtime"
)

// serveQuickListener serves the UDP listener of the VsoaServer.
//
// It takes an address string as a parameter and returns an error.
func (s *VsoaServer) serveQuickListener(address string) (err error) {
	qAddrServer := (*net.UDPAddr)(s.ln.Addr().(*net.TCPAddr))
	qln, err := net.ListenUDP("udp", qAddrServer)
	if err != nil {
		log.Fatal(err)
	}
	defer qln.Close()

	for {
		buf := make([]byte, 1024)
		_, addr, err := qln.ReadFromUDP(buf)
		qAddr := addr.String()
		if err != nil {
			continue
		} else {
			if clientUid, ok := s.quickChannel[(qAddr)]; ok {
				if client, ok := s.activeClients[clientUid]; ok {
					if client.Authed {
						req := protocol.NewMessage()
						r := bytes.NewBuffer(buf)
						err = req.Decode(r)
						if err != nil {
							if errors.Is(err, io.EOF) {
								log.Printf("Vsoa client has closed this connection: %s", qln.RemoteAddr().String())
							}

							if s.HandleServiceError != nil {
								s.HandleServiceError(err)
							}
							return err
						}
						go s.processOneQuickRequest(req, clientUid)
					}
				}
			}
		}
	}

}

// processOneQuickRequest processes a single quick request.
//
// It takes in a req of type *protocol.Message and ClientUid of type uint32.
// It does not return anything.
func (s *VsoaServer) processOneQuickRequest(req *protocol.Message, ClientUid uint32) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			buf = buf[:runtime.Stack(buf, true)]

			log.Printf("failed to handle the request: %v, stacks: %s", r, buf)
		}
	}()

	res := protocol.NewMessage()

	if handle, ok := s.routeMap["DATAGRAME."+string(req.URL)]; ok {
		handle(req, res)
	} else if handle, ok := s.routeMap["DATAGRAME.DEFAULT"]; ok {
		handle(req, res)
	}
}
