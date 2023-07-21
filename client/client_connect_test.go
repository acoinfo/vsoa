package client

import (
	"flag"
	"go-vsoa/protocol"
	"go-vsoa/server"
	"testing"
)

var (
	addr = flag.String("addr", "localhost:3003", "server address")
)

func TestGoodConnect(t *testing.T) {
	startServer()
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	SrvInfo, err := c.Connect("vsoa", *addr)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("SrvInfo:", SrvInfo, "ClientUid:", c.GetUid())
	}
	defer c.Close()
}

func TestWorngPasswdConnect(t *testing.T) {
	startServer()
	flag.Parse()

	clientOption := Option{
		Password: "12346",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *addr)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusPassword)) {
			t.Log("passed Passwd Err test")
		} else {
			t.Fatal(err)
		}
	}
	defer c.Close()
}

func startServer() {
	// Init golang server
	serverOption := server.Option{
		Password: "123456",
	}
	s := server.NewServer("golang VSOA server", serverOption)

	// Register URL
	h := func(req, res *protocol.Message) {
		res.Param = req.Param
		res.Data = req.Data
	}
	s.AddRpcHandler("/a/b/c", protocol.RpcMethodGet, h)

	go func() {
		_ = s.Serve("tcp", "127.0.0.1:3003")
	}()
	//defer s.Close()
	// Done init golang server
}
