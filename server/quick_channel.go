package server

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"runtime"
	"strings"

	"github.com/acoinfo/vsoa/protocol"
)

// serveQuickListener serves the UDP listener of the VsoaServer.
//
// It takes an address string as a parameter and returns an error.
func (s *Server) serveQuickListener(_ string) (err error) {
	qAddrServer := (*net.UDPAddr)(s.ln.Addr().(*net.TCPAddr))
	s.qln, err = net.ListenUDP("udp", qAddrServer)
	if err != nil {
		return err
	}
	defer func() {
		if s.qln != nil {
			_ = s.qln.Close()
		}
	}()

	buf := make([]byte, 1024)

	for {
		n, addr, err := s.qln.ReadFromUDP(buf)
		if err != nil {
			if s.IsShutdown() || errors.Is(err, net.ErrClosed) ||
				strings.Contains(err.Error(), "use of closed network connection") {
				return ErrServerClosed
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return err
		}
		if addr == nil {
			continue
		}

		qAddr := addr.String()
		if clientUid, ok := s.quickChannel[qAddr]; ok {
			if client, ok := s.clients[clientUid]; ok {
				if client.Active {
					req := protocol.NewMessage()
					r := bytes.NewBuffer(buf[:n])
					err = req.Decode(r)
					if err != nil {
						if errors.Is(err, io.EOF) {
							if s.HandleServiceError == nil {
								log.Printf("Vsoa client[%d] has closed this connection: %s", clientUid, qAddr)
							}
						}

						if s.HandleServiceError != nil {
							s.HandleServiceError(clientUid, err)
						}
						return err
					}
					go s.processOneQuickRequest(req, clientUid)
				}
			}
		} else {
			continue
		}
	}

}

// processOneQuickRequest processes a single quick request.
//
// It takes in a req of type *protocol.Message and ClientUid of type uint32.
// It does not return anything.
func (s *Server) processOneQuickRequest(req *protocol.Message, _ uint32) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			buf = buf[:runtime.Stack(buf, true)]

			log.Printf("failed to handle the request: %v, stacks: %s", r, buf)
		}
	}()

	res := protocol.NewMessage()

	if sh, ok := s.routeMap["DATAGRAME."+string(req.URL)]; ok {
		sh.handler(req, res)
	} else if sh, ok := s.routeMap["DATAGRAME.DEFAULT"]; ok {
		sh.handler(req, res)
	}
}
