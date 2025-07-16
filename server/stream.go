package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/acoinfo/vsoa/protocol"
)

type ServerStream struct {
	Tunid uint16
	Ln    net.Listener
}

// NewServerStream creates a new Stream using tunid in res.
//
// It takes a pointer to a Server object, s, and a pointer to a protocol.Message object, res, as parameters.
// It returns a pointer to a ServerStream object, ss, and an error object, err.
func (s *Server) NewServerStream(res *protocol.Message) (ss *ServerStream, err error) {
	var ln net.Listener
	var n int

	//We do this only for avoid mac & windows firewall blocking
	if s != nil {
		n = strings.LastIndexByte(s.address, ':')
	} else {
		return nil, fmt.Errorf("nil server")
	}

	ip := s.address[:n]

	ln, err = net.Listen("tcp", ip+":0")

	if err != nil {
		return nil, err
	}

	tunid := uint16(ln.Addr().(*net.TCPAddr).Port)

	res.SetTunId(tunid)
	res.SetValidTunid()

	return &ServerStream{
		Tunid: tunid,
		Ln:    ln,
	}, nil
}

func (ss *ServerStream) ServeListener(pushBuf, receiveBuf *bytes.Buffer) {
	var tempDelay time.Duration

	conn, e := ss.Ln.Accept()

	streamDone := make(chan int)

	if e != nil {
		if ne, ok := e.(net.Error); ok && ne.Timeout() {
			if tempDelay == 0 {
				tempDelay = 5 * time.Millisecond
			} else {
				tempDelay *= 2
			}
			if max := 1 * time.Second; tempDelay > max {
				tempDelay = max
			}

			time.Sleep(tempDelay)

		}
	} else {
		go func() {
			ss.serveConnPush(conn, pushBuf)
			streamDone <- 1
		}()
		ss.serveConnReceive(conn, receiveBuf)
	}

	<-streamDone

	ss.Ln.Close()
}

func (ss *ServerStream) serveConnPush(conn net.Conn, pushBuf *bytes.Buffer) {
	io.CopyN(conn, pushBuf, int64(pushBuf.Len()))
}

func (ss *ServerStream) serveConnReceive(conn net.Conn, receiveBuf *bytes.Buffer) {
	io.Copy(receiveBuf, conn)
}
