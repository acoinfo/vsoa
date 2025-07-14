package client

import (
	"flag"
	"testing"
	"time"

	"github.com/acoinfo/go-vsoa/protocol"
)

var (
	pingecho_addr = flag.String("pingecho_addr", "localhost:3003", "server address")
)

func TestPingEchot(t *testing.T) {
	startServer()
	flag.Parse()

	// Do this to make sure the server is ready on slow machine
	time.Sleep(50 * time.Millisecond)

	clientOption := Option{
		Password:     "123456",
		PingInterval: 2,
		PingTimeout:  1,
		PingLost:     2,
	}

	c := NewClient(clientOption)
	SrvInfo, err := c.Connect("vsoa", *pingecho_addr)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("SrvInfo:", SrvInfo, "ClientUid:", c.GetUid())
	}
	defer c.Close()

	time.Sleep(10 * time.Second)

	req := protocol.NewMessage()
	reply, err := c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		if reply.SeqNo() < 5 {
			t.Fatal("PingEcho not sended")
		}
		t.Log("Seq:", reply.SeqNo(), "Param:", (reply.Param))
	}
}
