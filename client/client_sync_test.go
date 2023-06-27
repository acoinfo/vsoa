package client

import (
	"flag"
	"go-vsoa/protocol"
	"testing"
)

var (
	addr = flag.String("addr", "localhost:3002", "server address")
)

func TestConnect(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	err := c.Connect("tcp", *addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}

func TestWorngPasswdConnect(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "12346",
	}

	c := NewClient(clientOption)
	err := c.Connect("tcp", *addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()
	reply := protocol.NewMessage()

	reply, err = c.Call(protocol.TypeServInfo, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusPassword)) {
			t.Log("Seq:", reply.SeqNo(), "passed Passwd Err test")
		}
	} else {
		t.Fatal("Failed, passwd should be err")
	}

	if c.IsShutdown() {
		t.Log("Client shutdown normally")
	} else {
		t.Fatal("Failed to shutdown client")
	}
}

func TestGoodHandShackConnect(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	err := c.Connect("tcp", *addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()
	reply := protocol.NewMessage()

	reply, err = c.Call(protocol.TypeServInfo, protocol.RpcMethodGet, req)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Seq:", reply.SeqNo(), "SrvInfo:", protocol.DecodeServInfo(reply.Param))
	}
}

func TestDoubleHandShackConnect(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	err := c.Connect("tcp", *addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()
	reply := protocol.NewMessage()

	reply, err = c.Call(protocol.TypeServInfo, protocol.RpcMethodGet, req)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Seq:", reply.SeqNo(), "SrvInfo:", protocol.DecodeServInfo(reply.Param))
	}

	reply, err = c.Call(protocol.TypeServInfo, protocol.RpcMethodGet, req)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Seq:", reply.SeqNo(), "SrvInfo:", protocol.DecodeServInfo(reply.Param))
	}
}
