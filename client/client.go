package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"go-vsoa/protocol"
	"io"
	"log"
	"net"
	"sync"
	"time"
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
	ErrUnsupportedCodec = errors.New("unsupported codec")
)

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 16 * 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 16 * 1024
)

type seqKey struct{}

// RPCClient is interface that defines one client to call one server.
type VsoaClient interface {
	// connect & shack hand with VSOA server
	Connect(network, address string) error
	// async func for RPC call
	GoRpc(mt protocol.MessageType, URL string, serviceMethod protocol.RpcMessageType, param *json.RawMessage, data []byte, done chan *RpcCall) *RpcCall
	// sync func wait answer
	Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error
	// send raw message without any codec
	SendRaw(ctx context.Context, r *protocol.Message) (map[string]string, []byte, error)
	Close() error
	// Return Server URL/ip:port
	RemoteAddr() string

	RegisterServerMessageChan(ch chan<- *protocol.Message)
	UnregisterServerMessageChan()

	IsClosing() bool
	IsShutdown() bool

	GetConn() net.Conn
}

// Client represents a VSOA client. (For NOW it's only RPC&ServInfo)
type Client struct {
	option Option

	Conn  net.Conn
	QConn net.Conn
	r     *bufio.Reader
	// w    *bufio.Writer

	mutex    sync.Mutex // protects following
	seq      uint32
	pending  map[uint32]*RpcCall
	closing  bool // user has called Close
	shutdown bool // server has told us to stop

	ServerMessageChan chan<- *protocol.Message
}

// NewClient returns a new Client with the option.
func NewClient(option Option) *Client {
	return &Client{
		option: option,
	}
}

// Option contains all options for creating clients.
type Option struct {
	Password       string
	PingInterval   int
	PingTimeout    int
	PingLost       int
	ConnectTimeout time.Duration
	// TLSConfig for tcp and quic
	TLSConfig *tls.Config
}

// Call represents an active RPC.
type RpcCall struct {
	URL           string                  // The Server URL
	ServiceMethod protocol.RpcMessageType // The name of the service and method to call.
	Data          []byte
	Param         *json.RawMessage
	Reply         *protocol.Message
	Error         error         // After completion, the error status.
	Done          chan *RpcCall // Strobes when call is complete.
}

func (call *RpcCall) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
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
func (client *Client) GoRpc(mt protocol.MessageType, serviceMethod protocol.RpcMessageType, req *protocol.Message, reply *protocol.Message, done chan *RpcCall) *RpcCall {
	call := new(RpcCall)
	call.URL = string(req.URL)
	call.ServiceMethod = serviceMethod

	call.Param = &req.Param
	call.Data = req.Data

	call.Reply = reply
	if done == nil {
		done = make(chan *RpcCall, 10) // buffered.
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
		go client.sendSrvInfo(call) // Internal use
	case protocol.TypeRPC:
		go client.sendRpc(call)
	default:
		return call // We just return done
	}

	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(mt protocol.MessageType, serviceMethod protocol.RpcMessageType, req *protocol.Message) (*protocol.Message, error) {
	return client.call(mt, serviceMethod, req)
}

func (client *Client) call(mt protocol.MessageType, serviceMethod protocol.RpcMessageType, req *protocol.Message) (*protocol.Message, error) {
	reply := protocol.NewMessage()

	Done := client.GoRpc(mt, serviceMethod, req, reply, make(chan *RpcCall, 1)).Done
	var err error
	select {
	case call := <-Done:
		err = call.Error
		reply = call.Reply
	}
	return reply, err
}

// Client send SrvInfo message
// Internal use for
func (client *Client) sendSrvInfo(call *RpcCall) {
	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}

	if client.pending == nil {
		client.pending = make(map[uint32]*RpcCall)
	}

	m := &protocol.ServInfoReqParam{
		Password:     client.option.Password,
		PingInterval: client.option.PingInterval,
		PingTimeout:  client.option.PingInterval,
		PingLost:     client.option.PingInterval,
	}

	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	req := protocol.NewMessage()
	m.NewMessage(req)
	req.SetSeqNo(seq)

	_, err := client.Conn.Write(req.Encode())

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
	return
}

// Client send RPC message
func (client *Client) sendRpc(call *RpcCall) {
	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}

	if client.pending == nil {
		client.pending = make(map[uint32]*RpcCall)
	}

	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	req := protocol.NewMessage()
	req.SetMessageType(protocol.TypeRPC)
	req.SetMessageRpcMethod(call.ServiceMethod)
	req.SetSeqNo(seq)

	if call.Param != nil {
		req.Param = *call.Param
	}

	req.URL = []byte(call.URL)
	req.Param = *call.Param
	req.Data = call.Data

	_, err := client.Conn.Write(req.Encode())

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

var count int = 0

func (client *Client) input() {
	var err error

	for err == nil {
		res := protocol.NewMessage()

		err = res.Decode(client.r)
		if err != nil {
			break
		}

		seq := res.SeqNo()
		var call *RpcCall
		isServerMessage := (res.IsReply() == false)
		if !isServerMessage {
			client.mutex.Lock()
			call = client.pending[seq]
			delete(client.pending, seq)
			client.mutex.Unlock()
		}

		switch {
		case call == nil:
			if isServerMessage {
				if client.ServerMessageChan != nil {
					client.handleServerRequest(res)
				}
				continue
			}
		case res.StatusType() != protocol.StatusSuccess:
			// We've got an error response. Give this to the request
			call.Error = strErr(res.StatusTypeText())
			call.Reply = res
			call.done()
		case res.StatusType() == protocol.StatusPassword:
			// We've got Passwd error response. Shutdown client
			call.Error = strErr(res.StatusTypeText())
			break
		default:
			call.Reply = res
			call.done()
		}
	}

	// Terminate pending calls.
	// This is used for Subscribe in VSOA
	if client.ServerMessageChan != nil {
		req := protocol.NewMessage()
		req.SetMessageType(protocol.TypePublish)
		req.SetStatusType(protocol.StatusNoResponding)
		client.handleServerRequest(req)
	}

	client.mutex.Lock()
	client.Conn.Close()
	client.shutdown = true
	closing := client.closing
	if e, ok := err.(*net.OpError); ok {
		if e.Addr != nil || e.Err != nil {
			err = fmt.Errorf("net.OpError: %s", e.Err.Error())
		} else {
			err = errors.New("net.OpError")
		}

	}
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}

	client.mutex.Unlock()

	if err != nil && !closing {
		log.Printf("VSOA: client protocol error: %v", err)
	}
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
	client.mutex.Unlock()
	return err
}
