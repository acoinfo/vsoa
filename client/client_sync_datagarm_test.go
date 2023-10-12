package client

import (
	"encoding/json"
	"flag"
	"testing"
	"time"

	"gitee.com/sylixos/go-vsoa/protocol"
	"gitee.com/sylixos/go-vsoa/server"
)

var (
	datagram_addr = flag.String("datagram_addr", "localhost:3003", "server address")
)

type DatagramTestParam struct {
	Num int `json:"Test Num"`
}

// TestDatagram is a test function that sends a datagram to the server(TCP).
func TestDatagram(t *testing.T) {
	startDatagramServer(t)
	flag.Parse()

	// Do this to make sure the server is ready on slow machine
	time.Sleep(50 * time.Millisecond)

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *datagram_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()

	req.Param, _ = json.RawMessage(`{"Test Num":123}`).MarshalJSON()

	_, err = c.Call("/datagram", protocol.TypeDatagram, protocol.ChannelNormal, req)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Datagram send done")
	}

	// don't close too quick before server handle the Call
	time.Sleep(5 * time.Millisecond)
}

// TestDatagramQuick is a function that tests the datagram functionality in Quick Channel(UDP).
func TestDatagramQuick(t *testing.T) {
	startDatagramServer(t)
	flag.Parse()

	// Do this to make sure the server is ready on slow machine
	time.Sleep(50 * time.Millisecond)

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *datagram_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()

	req.Param, _ = json.RawMessage(`{"Test Num":123}`).MarshalJSON()

	_, err = c.Call("/datagramQuick", protocol.TypeDatagram, protocol.ChannelQuick, req)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Datagram send done")
	}

	// don't close too quick before server handle the Call
	time.Sleep(5 * time.Millisecond)
}

// startDatagramServer initializes a golang server and registers URL handlers.
//
// It takes a *testing.T parameter for logging purposes.
func startDatagramServer(t *testing.T) {
	// Init golang server
	serverOption := server.Option{
		Password: "123456",
	}
	s := server.NewServer("golang VSOA server", serverOption)

	// Register URL
	h := func(req, res *protocol.Message) {
		res.Param = req.Param
		t.Log("/datagram Handler:", "URL", string(req.URL), "Param:", string(res.Param))
	}
	s.OnDatagarm("/datagram", h)
	qh := func(req, res *protocol.Message) {
		res.Param = req.Param
		t.Log("/datagramQuick Handler:", "URL", string(req.URL), "Param:", string(res.Param))
	}
	s.OnDatagarm("/datagramQuick", qh)
	dh := func(req, res *protocol.Message) {
		res.Param = req.Param
		t.Log("Default Handler:", "URL", string(req.URL), "Param:", string(res.Param))
	}
	s.OnDatagarmDefault(dh)

	go func() {
		_ = s.Serve("127.0.0.1:3003")
	}()
	//defer s.Close()
	// Done init golang server
}
