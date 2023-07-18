package server

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"go-vsoa/protocol"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ErrServerClosed is returned by the Server's Serve, ListenAndServe after a call to Shutdown or Close.
var (
	ErrServerClosed  = errors.New("http: Server closed")
	ErrReqReachLimit = errors.New("request reached rate limit")
)

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 1024

	DefaultTimeout = 500 * time.Millisecond
)

// VSOA Server need to konw the client infos
// until TCP connection closed
type client struct {
	Uid  uint32 // Server send this to client, helps client send quick massage
	Conn net.Conn
	// Quick channel Conn is UDP conn so we can't close real client conn
	// But we have to know the QAddr to delete quickChannel map
	QAddr          *net.UDPAddr
	beforeServInfo bool
	Authed         bool
	Subscribes     map[string]bool // key: URL, value: If Subs
}

// Handler declares the signature of a function that can be bound to a Route.
type Handler func(req *protocol.Message, resp *protocol.Message)

// VsoaServer is the VSOA server that use TCP with UDP.
type VsoaServer struct {
	Name         string //Used for ServInfo
	option       Option
	ln           net.Listener
	readTimeout  time.Duration
	writeTimeout time.Duration

	routerMapMu sync.RWMutex
	routeMap    map[string]Handler

	mu            sync.RWMutex
	activeClients map[uint32]*client
	// When QuickChannel get RemoteAddr we need to use it to check if we have the activeClient
	quickChannel map[*net.UDPAddr]uint32
	clientsCount atomic.Uint32
	doneChan     chan struct{}

	inShutdown int32
	onShutdown []func(s *VsoaServer)
	onRestart  []func(s *VsoaServer)

	// TLSConfig for creating tls tcp connection.
	tlsConfig *tls.Config

	handlerMsgNum int32

	// HandleServiceError is used to get all service errors. You can use it write logs or others.
	HandleServiceError func(error)

	// ServerErrorFunc is a customized error handlers and you can use it to return customized error strings to clients.
	// If not set, it use err.Error()
	ServerErrorFunc func(res *protocol.Message, err error) string
}

// NewServer returns a server.
func NewServer(name string, so Option) *VsoaServer {
	s := &VsoaServer{
		Name:          name,
		option:        so,
		readTimeout:   DefaultTimeout,
		writeTimeout:  DefaultTimeout,
		quickChannel:  make(map[*net.UDPAddr]uint32),
		activeClients: make(map[uint32]*client),
		doneChan:      make(chan struct{}),
		routeMap:      make(map[string]Handler),
	}

	return s
}

// Serve starts and listens VSOA normal channel requests.
// TODO: we need to start listen Quick channel too!
// It is blocked until receiving connections from clients.
func (s *VsoaServer) Serve(network, address string) (err error) {
	var ln net.Listener
	ln, err = s.makeListener(network, address)
	if err != nil {
		return err
	}

	return s.serveListener(ln)
}

// serveListener accepts incoming connections on the Listener ln,
// creating a new service goroutine for each.
// The service goroutines read requests and then call services to reply to them.
func (s *VsoaServer) serveListener(ln net.Listener) error {
	var tempDelay time.Duration

	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	for {
		conn, e := ln.Accept()
		if e != nil {
			if s.isShutdown() {
				<-s.doneChan
				return ErrServerClosed
			}

			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				time.Sleep(tempDelay)
				continue
			}

			return e
		}
		tempDelay = 0
		if _, ok := conn.(*net.TCPConn); ok {
		}

		CUid := s.clientsCount.Add(1)

		s.mu.Lock()
		// We don't know the QuickChannel info yet.
		s.activeClients[CUid] = &client{
			Conn: conn,
			Uid:  CUid,
			// until we got ServInfo
			QAddr:          nil,
			beforeServInfo: true,
			Authed:         false,
		}
		s.mu.Unlock()

		go s.serveConn(conn, CUid)
	}
}

func (s *VsoaServer) sendResponse(res *protocol.Message, conn net.Conn) {
	// Do service method
	tmp, err := res.Encode(protocol.ChannelNormal)
	if err != nil {
		log.Panicln(err)
		return
	}

	if s.writeTimeout != 0 {
		conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))
	}
	conn.Write(tmp)
	protocol.PutData(&tmp)
}

// This is a session for one client, so we need to keep auth logic inside it
func (s *VsoaServer) serveConn(conn net.Conn, ClientUid uint32) {
	if s.isShutdown() {
		s.closeConn(ClientUid)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			ss := runtime.Stack(buf, false)
			if ss > size {
				ss = size
			}
			buf = buf[:ss]
		}

		// make sure all inflight requests are handled and all drained
		if s.isShutdown() {
			<-s.doneChan
		}

		// close TCP&UDPconn
		s.closeConn(ClientUid)
	}()

	if tlsConn, ok := conn.(*tls.Conn); ok {
		if d := s.readTimeout; d != 0 {
			conn.SetReadDeadline(time.Now().Add(d))
		}
		if d := s.writeTimeout; d != 0 {
			conn.SetWriteDeadline(time.Now().Add(d))
		}
		if err := tlsConn.Handshake(); err != nil {
			return
		}
	}

	r := bufio.NewReaderSize(conn, ReaderBuffsize)

	// read requests and handle it
	for {
		if s.isShutdown() {
			return
		}

		t0 := time.Now()
		if s.readTimeout != 0 {
			conn.SetReadDeadline(t0.Add(s.readTimeout))
		}

		// read a request from the underlying connection
		req := protocol.NewMessage()
		err := req.Decode(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("Vsoa client has closed this connection: %s", conn.RemoteAddr().String())
			}

			if s.HandleServiceError != nil {
				s.HandleServiceError(err)
			}
			return
		}

		if !req.IsPingEcho() && !req.IsServInfo() {
			if s.activeClients[ClientUid].Authed == false {
				// Close unauthed client
				log.Printf("auth failed for conn %s: %v", conn.RemoteAddr().String(), protocol.StatusText(protocol.StatusPassword))
				return
			}
		}

		go s.processOneRequest(req, conn, ClientUid)
	}
}

func (s *VsoaServer) processOneRequest(req *protocol.Message, conn net.Conn, ClientUid uint32) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			buf = buf[:runtime.Stack(buf, true)]

			log.Printf("failed to handle the request: %v, stacks: %s", r, buf)
		}
	}()

	atomic.AddInt32(&s.handlerMsgNum, 1)
	defer atomic.AddInt32(&s.handlerMsgNum, -1)

	res := protocol.NewMessage()

	if req.IsServInfo() {
		err := s.servInfoHandler(req, res, ClientUid)
		if err != nil {
			log.Printf("Failed to Auth Client: %d", ClientUid)
		}
		s.sendResponse(res, conn)
		return
	}

	res = req.CloneHeader()
	res.SetReply(true)

	if req.IsPingEcho() {
		s.sendResponse(res, conn)
		return
	}

	if !req.IsOneway() {
		// TODO: father URL logic
		if handle, ok := s.routeMap[string(req.URL)+"."+req.MessageRpcMethodText()]; ok {
			handle(req, res)
		} else {
			res.SetStatusType(protocol.StatusInvalidUrl)
		}

		s.sendResponse(res, conn)
	} else {
		// TODO: internal handle
	}
}

// this is using for handle VSOA default servInfo type protocol
func (s *VsoaServer) servInfoHandler(req *protocol.Message, resp *protocol.Message, ClientUid uint32) error {
	r := new(protocol.ServInfoResParam)
	r.Info = s.Name
	s.mu.Lock()
	s.activeClients[ClientUid].beforeServInfo = false
	s.mu.Unlock()
	if s.option.Password == "" {
		s.mu.Lock()
		s.activeClients[ClientUid].Authed = true
		// quick channel register
		if req.TunID() != 0 {
			qAddr := (*net.UDPAddr)(s.activeClients[ClientUid].Conn.RemoteAddr().(*net.TCPAddr))
			qAddr.Port = int(req.TunID())
			s.activeClients[ClientUid].QAddr = (qAddr)
			s.quickChannel[(qAddr)] = ClientUid
		}
		s.mu.Unlock()
		r.NewGoodMessage(protocol.ServInfoResAsString, resp, ClientUid)
	} else {
		infoParam := new(protocol.ServInfoReqParam)
		err := json.Unmarshal(req.Param, infoParam)
		if err != nil {
			return err
		} else {
			if infoParam.Password != s.option.Password {
				r.NewErrMessage(resp)
				// After this call server to close normal channel conn
				return protocol.ErrMessagePasswd
			} else {
				s.mu.Lock()
				s.activeClients[ClientUid].Authed = true
				// quick channel register
				if req.TunID() != 0 {
					qAddr := (*net.UDPAddr)(s.activeClients[ClientUid].Conn.RemoteAddr().(*net.TCPAddr))
					qAddr.Port = int(req.TunID())
					s.activeClients[ClientUid].QAddr = (qAddr)
					s.quickChannel[(qAddr)] = ClientUid
				}
				s.mu.Unlock()
				r.NewGoodMessage(protocol.ServInfoResAsString, resp, ClientUid)
			}
		}
	}

	// TODO: handle other client options like ping echo seting logic
	return nil
}

// TODO: add both GET&SET with same handler!
func (s *VsoaServer) AddRpcHandler(servicePath string, serviceMethod protocol.RpcMessageType, handler func(*protocol.Message, *protocol.Message)) {
	s.routerMapMu.Lock()
	defer s.routerMapMu.Unlock()
	s.routeMap[servicePath+"."+protocol.RpcMethodText(serviceMethod)] = handler
}

func (s *VsoaServer) isShutdown() bool {
	return atomic.LoadInt32(&s.inShutdown) == 1
}

func (s *VsoaServer) closeConn(ClientUid uint32) {
	c := s.activeClients[ClientUid]
	// Quick Channel connection needs to close by client
	c.Conn.Close()

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.quickChannel, s.activeClients[ClientUid].QAddr)
	delete(s.activeClients, ClientUid)
}

// Option contains all options for creating server.
type Option struct {
	Password string
	// TLSConfig for tcp and quic
	TLSConfig *tls.Config
}
