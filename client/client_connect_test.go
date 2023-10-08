package client

import (
	"encoding/json"
	"flag"
	"go-vsoa/position"
	"go-vsoa/protocol"
	"go-vsoa/server"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	addr             = flag.String("addr", "localhost:3003", "server address")
	vsoa_test_server = "vsoa_test_server"
	position_addr    = "localhost:6003"

	StartPositionOnce sync.Once
	StartServerOnce   sync.Once
)

func TestGoodConnect(t *testing.T) {
	StartServerOnce.Do(func() {
		startServer()
	})
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
	StartServerOnce.Do(func() {
		startServer()
	})
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

func TestConnectWithPosition(t *testing.T) {
	StartPositionOnce.Do(func() {
		startPosition()
	})
	StartServerOnce.Do(func() {
		startServer()
	})

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	err := c.SetPosition(position_addr)
	if err != nil {
		t.Fatal(err)
	}
	SrvInfo, err := c.Connect(Type_URL, vsoa_test_server)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("SrvInfo:", SrvInfo, "ClientUid:", c.GetUid())
	}
	defer c.Close()
}

func TestConnectWithPositionNotFound(t *testing.T) {
	StartPositionOnce.Do(func() {
		startPosition()
	})
	StartServerOnce.Do(func() {
		startServer()
	})

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	err := c.SetPosition(position_addr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Connect(Type_URL, "foo_server")
	if err == position.ErrLookUpTimeOut || err == position.ErrServerNotFound {
		t.Log(err)
	} else {
		t.Fatal(err)
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

	hs := func(req, res *protocol.Message) {
		res.Param = req.Param
		res.Data = req.Data
	}
	s.AddRpcHandler("/a/b/c", protocol.RpcMethodSet, hs)

	pubs := func(req, _ *protocol.Message) {
		req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-Publishing"}`).MarshalJSON()
	}
	s.AddPublisher("/p", 1*time.Second, pubs)

	qpubs := func(req, _ *protocol.Message) {
		req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-Publishing-Quick"}`).MarshalJSON()
	}
	s.AddPublisher("/p/q", 1*time.Second, qpubs)

	go func() {
		_ = s.Serve("127.0.0.1:3003")
	}()
	//defer s.Close()
	// Done init golang server
}

func startPosition() {
	pl := position.NewPositionList()
	pl.Add(*position.NewPosition(vsoa_test_server, 1, "127.0.0.1", 3003, false))

	go pl.ServePositionListener(net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 6003,
	})
}
