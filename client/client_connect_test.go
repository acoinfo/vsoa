package client

import (
	"encoding/json"
	"flag"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-sylixos/go-vsoa/position"
	"github.com/go-sylixos/go-vsoa/protocol"
	"github.com/go-sylixos/go-vsoa/server"
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
		// Do this to make sure the server is ready on slow machine
		time.Sleep(50 * time.Millisecond)
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
		// Do this to make sure the server is ready on slow machine
		time.Sleep(50 * time.Millisecond)
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
		// Do this to make sure the position server is ready on slow machine
		time.Sleep(50 * time.Millisecond)
	})
	StartServerOnce.Do(func() {
		startServer()
		// Do this to make sure the server is ready on slow machine
		time.Sleep(50 * time.Millisecond)
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
		// Do this to make sure the position server is ready on slow machine
		time.Sleep(50 * time.Millisecond)
	})
	StartServerOnce.Do(func() {
		startServer()
		// Do this to make sure the server is ready on slow machine
		time.Sleep(50 * time.Millisecond)
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
		AutoAuth: true,
	}
	s := server.NewServer("golang VSOA server", serverOption)

	// Register URL
	h := func(req, res *protocol.Message) {
		res.Param = req.Param
		res.Data = req.Data
	}
	s.On("/a/b/c", protocol.RpcMethodGet, h)

	hs := func(req, res *protocol.Message) {
		res.Param = req.Param
		res.Data = req.Data
	}
	s.On("/a/b/c", protocol.RpcMethodSet, hs)

	pubs := func(req, _ *protocol.Message) {
		req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-Publishing"}`).MarshalJSON()
	}
	s.Publish("/p", 1*time.Second, pubs)

	pubd := func(req, _ *protocol.Message) {
		req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-Publishing on /p/d/"}`).MarshalJSON()
	}
	s.Publish("/p/d/", 1*time.Second, pubd)

	pubdd := func(req, _ *protocol.Message) {
		req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-Publishing on /p/d/d"}`).MarshalJSON()
	}
	s.Publish("/p/d/d", 1*time.Second, pubdd)

	qpubs := func(req, _ *protocol.Message) {
		req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-Publishing-Quick"}`).MarshalJSON()
	}
	s.QuickPublish("/p/q", 1*time.Second, qpubs)

	trigger := make(chan struct{}, 100)
	i := 1
	rawpubs := func(req, _ *protocol.Message) {
		i++
		req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-RAW-Publishing No. ` + strconv.Itoa(i) + `"}`).MarshalJSON()
	}
	s.Publish("/raw/p", trigger, rawpubs)

	go func() {
		_ = s.Serve("127.0.0.1:3003")
	}()

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if s.TriggerPublisher("/raw/p") != nil {
				break
			}
		}
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
