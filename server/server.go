package server

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/acoinfo/go-vsoa/protocol"
)

// ErrServerClosed is returned by the Server's Serve, ListenAndServe after a call to Shutdown or Close.
var (
	ErrServerClosed         = errors.New("VSOA: Server closed")
	ErrServerAlreadyStarted = errors.New("VSOA: Server Already Started")
	ErrReqReachLimit        = errors.New("request reached rate limit")
	ErrNilHandler           = errors.New("nil handler")
	ErrNilPublishHandler    = errors.New("nil publish handler")
	ErrWrongPublishTriger   = errors.New("wrong publish triger")
	ErrNotRawPublishURL     = errors.New("not raw publish URL, does not need triger")
	ErrAlreadyRegistered    = errors.New("URL has been Registered")
)

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 1024

	DefaultTimeout = 5 * time.Minute
)

// VsoaServer is interface that defines one client to call one server.
type VsoaServer interface {
	// NewServer returns a vsoa server.
	NewServer(name string, so Option) *Server
	// Serve starts and listens VSOA normal & quick channel requests.
	Serve(address string) (err error)
	// Close closes VSOA server.
	Close() (err error)
	// Get the number of connected clients.
	Count() int
	// Set on client funcs when client connect on server.
	OnClient(handler func(connect bool) (authed bool, err error))
	// On adds an RPC handler to the VsoaServer.
	On(servicePath string, serviceMethod protocol.RpcMessageType,
		handler func(*protocol.Message, *protocol.Message)) (err error)
	// OnDatagram adds a DATAGRAME handler to the VsoaServer.
	OnDatagram(servicePath string,
		handler func(*protocol.Message, *protocol.Message)) (err error)
	// OnDatagram adds a default DATAGRAME handler to the VsoaServer.
	OnDatagramDefault(
		handler func(*protocol.Message, *protocol.Message)) (err error)
	// Publish adds a publisher to the VsoaServer.
	Publish(servicePath string,
		timeDriction time.Duration,
		pubs func(*protocol.Message, *protocol.Message)) (err error)
	// QuickPublish adds a quick channel publisher to the server.
	QuickPublish(servicePath string,
		timeDriction time.Duration,
		pubs func(*protocol.Message, *protocol.Message)) (err error)
	// NewServerStream creates a new Stream using tunid in res.
	NewServerStream(res *protocol.Message) (ss *ServerStream, err error)
}

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
	Active         bool
	Subscribes     map[string]bool // key: URL, value: If Subs
}

// Handler declares the signature of a function that can be bound to a Route.
type Handler func(req *protocol.Message, resp *protocol.Message)
type serverHandler struct {
	handler Handler
	rawFlag bool
}

// Server is the VSOA server that use TCP with UDP.
type Server struct {
	Name         string //Used for ServInfo
	address      string
	option       Option
	ln           net.Listener
	qln          *net.UDPConn
	readTimeout  time.Duration
	writeTimeout time.Duration

	routerMapMu sync.RWMutex
	routeMap    map[string]serverHandler
	triggerChan map[string]chan struct{}

	mu      sync.RWMutex
	clients map[uint32]*client
	// When QuickChannel get RemoteAddr we need to use it to check if we have the activeClient
	quickChannel map[string]uint32
	clientsCount atomic.Uint32
	doneChan     chan struct{}

	isStarted  atomic.Bool
	isShutdown atomic.Bool
	// onShutdown []func(s *VsoaServer)
	// onRestart  []func(s *VsoaServer)

	// TLSConfig for creating tls tcp connection.
	tlsConfig *tls.Config

	handlerMsgNum int32

	// HandleServiceError is used to get all service errors. You can use it write logs or others.
	// This can be use to handler client close event.
	HandleServiceError func(clientUid uint32, err error)

	// HandleOnClient is used to do things if you want to do when client connect and active.
	// the return value `authed` is to make pubs to or not to goto the client.
	HandleOnClient func(clientUid uint32) (authed bool, err error)

	// ServerErrorFunc is a customized error handlers and you can use it to return customized error strings to clients.
	// If not set, it use err.Error()
	ServerErrorFunc func(res *protocol.Message, err error) string
}

// NewServer returns a server.
func NewServer(name string, so Option) *Server {
	if name == "" {
		name = "default GO-VSOA server name"
	}
	s := &Server{
		Name:   name,
		option: so,
		// this can cause server close connection
		readTimeout:  DefaultTimeout,
		writeTimeout: DefaultTimeout,
		quickChannel: make(map[string]uint32),
		clients:      make(map[uint32]*client),
		doneChan:     make(chan struct{}),
		routeMap:     make(map[string]serverHandler),
		triggerChan:  make(map[string]chan struct{}),
	}

	s.isStarted.Store(false)
	s.isShutdown.Store(false)

	return s
}

// Serve starts and listens VSOA normal & quick channel requests.
// It is blocked until receiving connections from clients.
func (s *Server) Serve(address string) (err error) {
	if s.IsStarted() {
		return ErrServerAlreadyStarted
	}

	var ln net.Listener
	ln, err = s.makeListener("tcp", address)
	if err != nil {
		return err
	}

	s.address = address
	s.isStarted.Store(true)
	s.isShutdown.Store(false)

	// Go quick channel listener
	go s.serveQuickListener(address)

	return s.serveListener(ln)
}

func (s *Server) IsStarted() bool {
	return s.isStarted.Load()
}

func (s *Server) Close() (err error) {
	if !s.isStarted.Load() || s.IsShutdown() {
		return nil
	}

	for cuid := range s.clients {
		s.closeConn(cuid)
	}

	s.mu.Lock()
	s.ln.Close()
	s.qln.Close()
	s.mu.Unlock()

	s.isStarted.Store(false)
	s.isShutdown.Store(true)

	return nil
}

func (s *Server) Count() (count int) {
	if !s.isStarted.Load() || s.IsShutdown() {
		return 0
	}

	count = 0

	s.mu.Lock()
	defer s.mu.Unlock()

	for cuid := range s.clients {
		if s.clients[cuid].Active {
			count++
		}
	}
	return count
}

// serveListener accepts incoming connections on the Listener ln,
// creating a new service goroutine for each.
// The service goroutines read requests and then call services to reply to them.
func (s *Server) serveListener(ln net.Listener) error {
	var tempDelay time.Duration

	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	for {
		conn, e := s.ln.Accept()
		if e != nil {
			if s.IsShutdown() {
				<-s.doneChan
				return ErrServerClosed
			}

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
				continue
			}

			return e
		}
		tempDelay = 0

		CUid := s.clientsCount.Add(1)

		s.mu.Lock()
		// We don't know the QuickChannel info yet.
		s.clients[CUid] = &client{
			Conn: conn,
			Uid:  CUid,
			// until we got ServInfo
			QAddr:          nil,
			beforeServInfo: true,
			Active:         false,
			Authed:         false,
		}
		s.mu.Unlock()

		go s.serveConn(conn, CUid)
	}
}

// sendResponse sends a response to the client.
//
// It takes in a res *protocol.Message object, representing the response to be sent,
// and a conn net.Conn object, representing the network connection to the client.
// The function encodes the response message and writes it to the connection.
// If a write timeout is set, it sets the write deadline before writing the response.
// The function returns no values.
func (s *Server) sendResponse(res *protocol.Message, conn net.Conn) {
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

// serveConn serves a connection and handles incoming requests for the VsoaServer.
// This is a session for one client, so we need to keep auth logic inside it
//
// Parameters:
// - conn: the net.Conn representing the connection.
// - ClientUid: the uint32 representing the client's UID.
//
// Returns: none.
func (s *Server) serveConn(conn net.Conn, ClientUid uint32) {
	if s.IsShutdown() {
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
			log.Printf("serving %s panic error: %s, stack:\n %s", conn.RemoteAddr(), err, buf)
		}

		// make sure all inflight requests are handled and all drained
		if s.IsShutdown() {
			<-s.doneChan
		}
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
		if s.IsShutdown() {
			return
		}

		t0 := time.Now()
		// If client send nothing during readTimeout to server, server will kill the connection!
		if s.readTimeout != 0 {
			conn.SetReadDeadline(t0.Add(s.readTimeout))
		}

		// read a request from the underlying connection
		req := protocol.NewMessage()
		// If client send nothing during readTimeout to server, can cause error
		err := req.Decode(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if s.HandleServiceError == nil {
					log.Printf("Vsoa client[%d] has closed this connection: %s", ClientUid, conn.RemoteAddr().String())
				}
			}

			if s.HandleServiceError != nil {
				s.HandleServiceError(ClientUid, err)
			}
			return
		}

		if !req.IsServInfo() {
			if !s.clients[ClientUid].Active {
				// Close unauthed client
				s.closeConn(ClientUid)
				log.Printf("auth failed for conn %s: %v", conn.RemoteAddr().String(), protocol.StatusText(protocol.StatusPassword))
				return
			}
		}

		go s.processOneRequest(req, conn, ClientUid)
	}
}

// processOneRequest is a method of the VsoaServer struct that processes a single request.
//
// It takes the following parameters:
// - req: a pointer to a protocol.Message struct, representing the request message
// - conn: a net.Conn object, representing the network connection
// - ClientUid: an unsigned 32-bit integer, representing the client UID
//
// There is no return value.
func (s *Server) processOneRequest(req *protocol.Message, conn net.Conn, ClientUid uint32) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			buf = buf[:runtime.Stack(buf, true)]

			log.Printf("failed to handle the request: %v, stacks: %s", r, buf)
		}
	}()

	atomic.AddInt32(&s.handlerMsgNum, 1)
	defer atomic.AddInt32(&s.handlerMsgNum, -1)

	res := req.CloneHeader()

	if req.IsServInfo() {
		err := s.servInfoHandler(req, res, ClientUid)
		if err != nil {
			log.Printf("Failed to Auth Client: %d, err: %s", ClientUid, err)
		}
		s.sendResponse(res, conn)
		return
	}

	res.SetReply(true)

	if req.IsNoop() {
		res.Reset()
		return
	}

	if req.IsPingEcho() {
		s.sendResponse(res, conn)
		return
	}

	if !req.IsOneway() {
		if req.IsRPC() {
			if sh, ok := s.routeMap["RPC."+req.MessageRpcMethodText()+
				"."+string(req.URL)]; ok {
				if sh.handler != nil {
					sh.handler(req, res)
				}
				res.SetStatusType(protocol.StatusSuccess)
				goto SEND
			}
			if sh, ok := s.routeMap["RPC."+req.MessageRpcMethodText()+
				"."+string(req.URL)+"/"]; ok {
				if sh.handler != nil {
					sh.handler(req, res)
				}
				res.SetStatusType(protocol.StatusSuccess)
				goto SEND
			}
			// wdie check if any matches
			for route, sh := range s.routeMap {
				// Find one handler and send
				if strings.HasSuffix(route, "/") &&
					strings.HasPrefix("RPC."+
						req.MessageRpcMethodText()+
						"."+string(req.URL), route) {
					if sh.handler != nil {
						sh.handler(req, res)
					}
					res.SetStatusType(protocol.StatusSuccess)
					goto SEND
				}
			}
			res.SetStatusType(protocol.StatusInvalidUrl)
			goto SEND
		} else if req.IsSubscribe() || req.IsUnSubscribe() {
			if _, ok := s.routeMap["SUBS/UNSUBS."+string(req.URL)]; ok &&
				!strings.HasSuffix(string(req.URL), "/") {
				s.subs(req, ClientUid)
				res.SetStatusType(protocol.StatusSuccess)
				goto SEND
			}
			if _, ok := s.routeMap["SUBS/UNSUBS."+string(req.URL)+"/"]; ok &&
				!strings.HasSuffix(string(req.URL), "/") {
				s.subsF(req, ClientUid)
				res.SetStatusType(protocol.StatusSuccess)
				goto SEND
			}
			if _, ok := s.routeMap["SUBS/UNSUBS."+string(req.URL)]; ok &&
				strings.HasSuffix(string(req.URL), "/") {
				for route := range s.routeMap {
					if strings.HasPrefix(route, "SUBS/UNSUBS."+string(req.URL)) {
						s.subsURL(req, route[12:], ClientUid)
					}
				}
				res.SetStatusType(protocol.StatusSuccess)
				goto SEND
			}
			if _, ok := s.routeMap["SUBS/UNSUBS."+string(req.URL[:len(req.URL)-1])]; ok &&
				strings.HasSuffix(string(req.URL), "/") {
				s.subsS(req, ClientUid)
				res.SetStatusType(protocol.StatusSuccess)
				goto SEND
			}
			res.SetStatusType(protocol.StatusInvalidUrl)
			goto SEND
		} else {
			res.SetStatusType(protocol.StatusInvalidUrl)
		}

	SEND:
		s.sendResponse(res, conn)
	} else {
		if sh, ok := s.routeMap["DATAGRAME."+string(req.URL)]; ok {
			if sh.handler != nil {
				sh.handler(req, res)
			}
			return
		}
		if sh, ok := s.routeMap["DATAGRAME."+req.MessageRpcMethodText()+
			"."+string(req.URL)+"/"]; ok {
			if sh.handler != nil {
				sh.handler(req, res)
			}
			return
		}
		// wdie check if any matches
		for route, sh := range s.routeMap {
			// Find one and return
			if strings.HasSuffix(route, "/") && strings.HasPrefix("DATAGRAME."+string(req.URL), route) {
				if sh.handler != nil {
					sh.handler(req, res)
				}
				return
			}
		}
		// We still have a Default here
		if sh, ok := s.routeMap["DATAGRAME.DEFAULT"]; ok {
			if sh.handler != nil {
				sh.handler(req, res)
			}
		}
	}
}

// servInfoHandler handles the server information request from a client.
//
// It takes in the request message, the response message, and the client UID as parameters.
// It returns an error if any error occurs during the process.
func (s *Server) servInfoHandler(req *protocol.Message, resp *protocol.Message, ClientUid uint32) error {
	r := new(protocol.ServInfoResParam)
	r.Info = s.Name
	s.mu.Lock()
	s.clients[ClientUid].beforeServInfo = false
	s.mu.Unlock()
	if s.option.Password == "" {
		s.mu.Lock()
		s.clients[ClientUid].Active = true
		// quick channel register
		if req.TunID() != 0 {
			qAddr := (*net.UDPAddr)(s.clients[ClientUid].Conn.RemoteAddr().(*net.TCPAddr))
			qAddr.Port = int(req.TunID())
			qString := qAddr.String()
			s.clients[ClientUid].QAddr = (qAddr)
			s.quickChannel[(qString)] = ClientUid
		}
		s.clients[ClientUid].Subscribes = make(map[string]bool)
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
				s.clients[ClientUid].Active = true
				// quick channel register
				if req.TunID() != 0 {
					qAddr := (*net.UDPAddr)(s.clients[ClientUid].Conn.RemoteAddr().(*net.TCPAddr))
					qAddr.Port = int(req.TunID())
					qString := qAddr.String()
					s.clients[ClientUid].QAddr = (qAddr)
					s.quickChannel[(qString)] = ClientUid
				}
				s.clients[ClientUid].Subscribes = make(map[string]bool)
				s.mu.Unlock()
				r.NewGoodMessage(protocol.ServInfoResAsString, resp, ClientUid)
			}
		}
	}

	s.onClient(ClientUid)
	// TODO: handle other client options like ping echo seting logic
	return nil
}

func (s *Server) OnClient(handler func(clientUid uint32) (authed bool, err error)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.HandleOnClient = handler
}

func (s *Server) onClient(ClientUid uint32) (err error) {
	if s.HandleOnClient == nil {
		s.authClient(ClientUid, s.option.AutoAuth)
		return defaultOnClientHandler(ClientUid)
	} else {
		authed, err := s.HandleOnClient(ClientUid)
		s.authClient(ClientUid, authed)
		return err
	}
}

func (s *Server) authClient(ClientUid uint32, authed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[ClientUid].Authed = authed
}

func defaultOnClientHandler(clientUid uint32) (err error) {
	log.Printf("Vsoa client[%d] connected!", clientUid)
	return nil
}

// TODO: add both GET&SET with same handler!
// On adds an RPC handler to the VsoaServer.
//
// It takes in the following parameters:
// - servicePath: the path of the service
// - serviceMethod: the type of the RPC message
// - handler: the function to handle the RPC message
//
// It returns an error.
func (s *Server) On(servicePath string, serviceMethod protocol.RpcMessageType, handler func(*protocol.Message, *protocol.Message)) (err error) {
	if handler == nil {
		return ErrNilHandler
	}
	s.routerMapMu.Lock()
	defer s.routerMapMu.Unlock()
	if _, ok := s.routeMap["RPC."+protocol.RpcMethodText(serviceMethod)+"."+servicePath]; !ok {
		s.routeMap["RPC."+protocol.RpcMethodText(serviceMethod)+"."+servicePath] = serverHandler{handler: handler, rawFlag: false}
	} else {
		return ErrAlreadyRegistered
	}
	return nil
}

// OnDatagram adds a DATAGRAME handler to the VsoaServer.
//
// It takes in the servicePath string and the handler function, and returns an error.
func (s *Server) OnDatagram(servicePath string, handler func(*protocol.Message, *protocol.Message)) (err error) {
	if handler == nil {
		return ErrNilHandler
	}
	s.routerMapMu.Lock()
	defer s.routerMapMu.Unlock()
	if _, ok := s.routeMap["DATAGRAME."+servicePath]; !ok {
		s.routeMap["DATAGRAME."+servicePath] = serverHandler{handler: handler, rawFlag: false}
	} else {
		return ErrAlreadyRegistered
	}
	return nil
}

// OnDatagramDefault adds a default DATAGRAME handler to the VsoaServer.
//
// The handler parameter is a function that takes two parameters: a pointer to a protocol.Message
// and a pointer to another protocol.Message. It is responsible for handling the ondata event.
// This function does not return anything.
func (s *Server) OnDatagramDefault(handler func(*protocol.Message, *protocol.Message)) (err error) {
	if handler == nil {
		return ErrNilHandler
	}
	s.routerMapMu.Lock()
	defer s.routerMapMu.Unlock()
	s.routeMap["DATAGRAME.DEFAULT"] = serverHandler{handler: handler, rawFlag: false}
	return nil
}

// Publish adds a publisher to the VsoaServer.
// Use this function to register a publish handler after Subs calls.
//
// The function takes the following parameter(s):
// - servicePath: a string representing the service path
// - timeOrTrigger: a time duration representing the time duration or a trigger to send pubs in raw ways.
// - pubs: a function that takes two pointers to protocol.Message and returns nothing
//
// It returns an error.
func (s *Server) Publish(servicePath string, timeOrTrigger any, pubs func(*protocol.Message, *protocol.Message)) (err error) {
	if pubs == nil {
		return ErrNilPublishHandler
	}
	rawFlag := false
	switch timeOrTrigger.(type) {
	case time.Duration:
	case chan struct{}:
		rawFlag = true
	default:
		return ErrWrongPublishTriger
	}
	s.routerMapMu.Lock()
	defer s.routerMapMu.Unlock()

	if _, ok := s.routeMap["SUBS/UNSUBS."+servicePath]; !ok {
		// No need to have handler save in the routeMap
		s.routeMap["SUBS/UNSUBS."+servicePath] = serverHandler{handler: pubs, rawFlag: rawFlag}
		// Maybe it's bad to run a Publisher for each pub
		go s.publisher(servicePath, timeOrTrigger, pubs)
	} else {
		return ErrAlreadyRegistered
	}
	return nil
}

// QuickPublish adds a quick channel publisher to the server.
//
// Parameters:
// - servicePath: the path of the service
// - timeOrTrigger: a time duration representing the time duration or a trigger to send pubs in raw ways.
// - pubs: a function that takes two protocol.Message parameters and returns nothing
//
// Returns:
// - err: an error if the publisher is already registered, otherwise nil
func (s *Server) QuickPublish(servicePath string, timeOrTrigger any, pubs func(*protocol.Message, *protocol.Message)) (err error) {
	if pubs == nil {
		return ErrNilPublishHandler
	}
	rawFlag := false
	switch timeOrTrigger.(type) {
	case time.Duration:
	case chan struct{}:
		rawFlag = true
	default:
		return ErrWrongPublishTriger
	}
	s.routerMapMu.Lock()
	defer s.routerMapMu.Unlock()

	if _, ok := s.routeMap["SUBS/UNSUBS."+servicePath]; !ok {
		// No need to have handler save in the routeMap
		s.routeMap["SUBS/UNSUBS."+servicePath] = serverHandler{handler: pubs, rawFlag: rawFlag}
		// Maybe it's bad to run a Publisher for each pub
		go s.qpublisher(servicePath, timeOrTrigger, pubs)
	} else {
		return ErrAlreadyRegistered
	}
	return nil
}

func (s *Server) TriggerPublisher(servicePath string) error {
	if s.routeMap["SUBS/UNSUBS."+servicePath].handler == nil {
		return ErrNilPublishHandler
	}

	if !s.routeMap["SUBS/UNSUBS."+servicePath].rawFlag {
		return ErrNotRawPublishURL
	}

	if s.triggerChan[servicePath] == nil {
		s.triggerChan[servicePath] = make(chan struct{}, 100)
	}

	s.triggerChan[servicePath] <- struct{}{}
	return nil
}

// subs updates the subscription status of a client.
//
// It takes a request message and a client UID as parameters and updates the
// subscription status of the client accordingly. If the request is a subscribe
// request, it sets the subscription status of the client to true for the given
// URL. Otherwise, it sets the subscription status to false.
func (s *Server) subs(req *protocol.Message, ClientUid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.IsSubscribe() {
		s.clients[ClientUid].Subscribes[string(req.URL)] = true
	} else {
		s.clients[ClientUid].Subscribes[string(req.URL)] = false
	}
}

func (s *Server) subsURL(req *protocol.Message, URL string, ClientUid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.IsSubscribe() {
		s.clients[ClientUid].Subscribes[URL] = true
	} else {
		s.clients[ClientUid].Subscribes[URL] = false
	}
}

func (s *Server) subsF(req *protocol.Message, ClientUid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.IsSubscribe() {
		s.clients[ClientUid].Subscribes[string(req.URL)+"/"] = true
	} else {
		s.clients[ClientUid].Subscribes[string(req.URL)+"/"] = false
	}
}

func (s *Server) subsS(req *protocol.Message, ClientUid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.IsSubscribe() {
		s.clients[ClientUid].Subscribes[string(req.URL[:len(req.URL)-1])] = true
	} else {
		s.clients[ClientUid].Subscribes[string(req.URL[:len(req.URL)-1])] = false
	}
}

// IsShutdown checks if the VsoaServer is in the shutdown state.
//
// It returns a boolean value indicating whether the VsoaServer is in the
// shutdown state or not.
func (s *Server) IsShutdown() bool {
	return s.isShutdown.Load()
}

// closeConn closes the connection for a given client in the VsoaServer.
//
// It takes the ClientUid as a parameter, which is the unique identifier of the client.
// There is no return type for this function.
func (s *Server) closeConn(ClientUid uint32) {
	c := s.clients[ClientUid]
	// Quick Channel connection needs to close by client
	c.Conn.Close()

	s.mu.Lock()
	defer s.mu.Unlock()
	// Clear Client Subscribes map
	for k := range s.clients[ClientUid].Subscribes {
		delete(s.clients[ClientUid].Subscribes, k)
	}
	delete(s.quickChannel, s.clients[ClientUid].QAddr.String())
	delete(s.clients, ClientUid)
}

// Option contains all options for creating server.
type Option struct {
	Password string
	// TLSConfig for tcp and quic
	TLSConfig *tls.Config
	// automatic auth all clients to get pubs
	AutoAuth bool
}
