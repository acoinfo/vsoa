package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"gitee.com/sylixos/go-vsoa/protocol"
)

// ServiceError is an error from server.
type ServiceError interface {
	Error() string
	IsServiceError() bool
}

var ClientErrorFunc func(e string) ServiceError

type strErr string

func (s strErr) Error() string {
	return string(s)
}

func (s strErr) IsServiceError() bool {
	return true
}

// DefaultOption is a common option configuration for client.
var DefaultOption = Option{}

// ErrShutdown connection is closed.
var (
	ErrShutdown         = errors.New("connection is shut down")
	ErrUnAuthed         = errors.New("client is not Authed")
	ErrUnsupportedCodec = errors.New("unsupported codec")
	ErrPingEcho         = errors.New("PingEcho set error")
)

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 16 * 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 16 * 1024
)

// VsoaClient is interface that defines one client to call one server.
type VsoaClient interface {
	// connect & shack hand with VSOA server
	Connect(network, address string) error
	// async func for VSOA call
	Go(mt protocol.MessageType, URL string, serviceMethod protocol.RpcMessageType, param *json.RawMessage, data []byte, done chan *Call) *Call
	// sync func for VSOA call
	Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error

	// Close the client & release the resources
	Close() error
	// Return Server URL/ip:port
	RemoteAddr() string

	// Subscribe server URL;
	// inject onPublish callback to the URL with father URL
	Subscribe(URL string, onPublish func(m *protocol.Message)) error

	// UnSubscribe server URL;
	// delete onPublish callback to the URL with father URL
	UnSubscribe(URL string) error

	// If server authed this client
	IsAuthed() bool
	// If client is closing or not
	IsClosing() bool
	// If client is shutdown or not
	IsShutdown() bool
}

// Client represents a VSOA client. (For NOW it's only RPC&ServInfo)
type Client struct {
	addr     string
	position string
	option   Option
	uid      uint32

	Conn net.Conn
	r    *bufio.Reader
	// Quick Datagram/Publish goes UDPs
	QConn *net.UDPConn
	qr    *bufio.Reader

	// used for server publish
	SubscribeList map[string]func(m *protocol.Message)

	mutex sync.Mutex // protects following

	noseq            uint32
	seq              uint32
	pending          map[uint32]*Call
	authed           bool  // if server authed this client
	closing          bool  // user has called Close
	shutdown         bool  // server has told us to stop
	pingTimeoutCount int32 // for server ping echo logic

	ServerMessageChan chan<- *protocol.Message
}

// NewClient returns a new Client with the option.
func NewClient(option Option) *Client {
	if option.ConnectTimeout == 0 {
		option.ConnectTimeout = 5 * time.Second
	}
	return &Client{
		option:           option,
		authed:           false,
		pingTimeoutCount: 0,
	}
}

func (c *Client) clearClient() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.authed = false
	c.closing = false
	c.shutdown = false
	c.pingTimeoutCount = 0
}

func (c *Client) GetUid() uint32 {
	return c.uid
}

// Option contains all options for creating clients.
type Option struct {
	Password       string
	PingInterval   int //second
	PingTimeout    int
	PingLost       int32
	PingTurbo      int //millisecond
	ConnectTimeout time.Duration
	// TLSConfig for tcp and quic
	TLSConfig *tls.Config
}

// Call represents an active RPC.
type Call struct {
	URL           string                    // The Server URL
	VsoaType      protocol.MessageType      // The real method when VSOA call out
	ServiceMethod protocol.RpcMessageType   // The name of the service and method to call.
	IsQuick       protocol.QuickChannelFlag //For Datagram/Publish to kown it's UDP channel or not
	Data          []byte
	Param         *json.RawMessage
	Reply         *protocol.Message
	Error         error      // After completion, the error status.
	Done          chan *Call // Strobes when call is complete.
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

// IsAuthed client is closing or not.
func (client *Client) IsAuthed() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.authed
}

// IsClosing client is closing or not.
func (client *Client) IsClosing() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.closing
}

// IsShutdown client is shutdown or not.
func (client *Client) IsShutdown() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.shutdown
}

// Go invokes the function asynchronously. It returns the Call structure representing
// the invocation. The done channel will signal when the call is complete by returning
// the same Call object. If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(URL string, mt protocol.MessageType, flags any, req *protocol.Message, reply *protocol.Message, done chan *Call) *Call {
	call := new(Call)
	call.URL = URL // prase the URL when go func "send"

	call.IsQuick = false
	call.ServiceMethod = protocol.RpcMethodGet

	switch t := flags.(type) {
	case protocol.RpcMessageType:
		if mt == protocol.TypeRPC {
			call.ServiceMethod = t
		}
	case protocol.QuickChannelFlag:
		if mt == protocol.TypeDatagram || mt == protocol.TypePublish {
			// This is used for UDP Quick chennels
			call.IsQuick = t
		}
	default:
		// Do nothing
	}

	call.Param = &req.Param
	call.Data = req.Data

	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done

	switch mt {
	case protocol.TypeServInfo:
		go client.sendSrvInfo(call) // Internal use mostly, But still user can call it,
	case protocol.TypeRPC:
		go client.sendRPC(call)
	case protocol.TypeDatagram:
		go client.sendSingle(call)
	case protocol.TypeSubscribe:
		go client.sendSubscribe(call, true)
	case protocol.TypeUnsubscribe:
		go client.sendSubscribe(call, false)
	case protocol.TypeNoop:
		go client.sendNoop(call)
	case protocol.TypePingEcho:
		go client.sendPingEcho(call)
	default:
		return call // We just return done
	}

	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(URL string, mt protocol.MessageType, flags any, req *protocol.Message) (*protocol.Message, error) {
	return client.call(URL, mt, flags, req)
}

func (client *Client) call(URL string, mt protocol.MessageType, flags any, req *protocol.Message) (*protocol.Message, error) {
	reply := protocol.NewMessage()

	Done := client.Go(URL, mt, flags, req, reply, make(chan *Call, 1)).Done
	var err error

	call := <-Done
	err = call.Error
	reply = call.Reply

	return reply, err
}

// Client send SrvInfo message
// Internal use for handshake with server.
func (client *Client) sendSrvInfo(call *Call) {
	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}

	if client.pending == nil {
		client.pending = make(map[uint32]*Call)
	}

	m := &protocol.ServInfoReqParam{
		Password:     client.option.Password,
		PingInterval: client.option.PingInterval,
		PingTimeout:  client.option.PingTimeout,
		PingLost:     client.option.PingLost,
	}

	seq := client.seq
	client.seq++
	client.pending[seq] = call

	req := protocol.NewMessage()
	if client.QConn == nil {
		// It should not be here
		m.NewMessage(req, "127.0.0.1:60000")
	} else {
		m.NewMessage(req, client.QConn.LocalAddr().String())
	}
	req.SetSeqNo(seq)

	tmp, err := req.Encode(protocol.ChannelNormal)
	if err != nil {
		call = client.pending[seq]
		delete(client.pending, seq)
		call.Error = err
		client.mutex.Unlock()
		call.done()
		return
	}
	client.mutex.Unlock()

	_, err = client.Conn.Write(tmp)

	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e.Err != nil {
				err = fmt.Errorf("net.OpError: %s", e.Err.Error())
			} else {
				err = errors.New("net.OpError")
			}

		}
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}
	// We don't done the Call, util we get Server Input
}

// Client send RPC message
func (client *Client) sendRPC(call *Call) {
	// If it's RPC call Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}

	if client.pending == nil {
		client.pending = make(map[uint32]*Call)
	}

	seq := client.seq
	client.seq++
	client.pending[seq] = call

	req := protocol.NewMessage()
	req.SetMessageType(protocol.TypeRPC)
	req.SetMessageRpcMethod(call.ServiceMethod)
	req.SetSeqNo(seq)

	req.URL = []byte(call.URL)
	req.Param = *call.Param
	req.Data = call.Data

	tmp, err := req.Encode(protocol.ChannelNormal)
	if err != nil {
		call = client.pending[seq]
		delete(client.pending, seq)
		call.Error = err
		client.mutex.Unlock()
		call.done()
		return
	}
	client.mutex.Unlock()

	_, err = client.Conn.Write(tmp)

	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e.Err != nil {
				err = fmt.Errorf("net.OpError: %s", e.Err.Error())
			} else {
				err = errors.New("net.OpError")
			}

		}
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}
}

// Client send Datagram(TCP/UDP) message
func (client *Client) sendSingle(call *Call) {
	// If it's Datagram call Set header's seq always be zero
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}

	req := protocol.NewMessage()
	req.SetMessageType(protocol.TypeDatagram)
	if call.IsQuick {
		// This is Quick channel specific
		req.SetSeqNo(client.uid)
	} else {
		req.SetSeqNo(0)
	}

	req.URL = []byte(call.URL)
	req.Param = *call.Param
	req.Data = call.Data

	tmp, err := req.Encode(call.IsQuick)
	if err != nil {
		call.Error = err
		client.mutex.Unlock()
		call.done()
		return
	}
	client.mutex.Unlock()

	if call.IsQuick {
		_, err = client.QConn.Write(tmp)
	} else {
		_, err = client.Conn.Write(tmp)
	}

	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e.Err != nil {
				err = fmt.Errorf("net.OpError: %s", e.Err.Error())
			} else {
				err = errors.New("net.OpError")
			}

		}
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}

	// Datagram don't have respond
	call.done()
}

func (client *Client) handleServerRequest(msg *protocol.Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ServerMessageChan may be closed so client remove it. Please add it again if you want to handle server requests. error is %v", r)
			client.ServerMessageChan = nil
		}
	}()

	serverMessageChan := client.ServerMessageChan
	if serverMessageChan != nil {
		select {
		case serverMessageChan <- msg:
		default:
			log.Panicf("ServerMessageChan may be full so the server request %d has been dropped", msg.SeqNo())
		}
	}
}

// Close calls the underlying connection's Close method. If the connection is already
// shutting down, ErrShutdown is returned.
func (client *Client) Close() error {
	client.mutex.Lock()

	for seq, call := range client.pending {
		delete(client.pending, seq)
		if call != nil {
			call.Error = ErrShutdown
			call.done()
		}
	}

	var err error
	if client.closing || client.shutdown {
		client.mutex.Unlock()
		return ErrShutdown
	}

	client.closing = true

	if client.QConn != nil {
		client.QConn.Close()
	}
	if client.Conn != nil {
		err = client.QConn.Close()
	}

	client.closing = false
	client.shutdown = true

	client.mutex.Unlock()

	return err
}
