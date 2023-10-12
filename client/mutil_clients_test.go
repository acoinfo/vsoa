package client

import (
	"encoding/json"
	"flag"
	"testing"
	"time"

	"gitee.com/sylixos/go-vsoa/protocol"
)

var (
	our_addr = flag.String("our_addr", "localhost:3003", "server address")
)

func TestMutilClientsConnect(t *testing.T) {
	startServer()
	flag.Parse()

	// Do this to make sure the server is ready on slow machine
	time.Sleep(50 * time.Millisecond)

	clientOption := Option{
		Password: "123456",
	}

	c1 := NewClient(clientOption)
	SrvInfo, err := c1.Connect("vsoa", *our_addr)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("SrvInfo:", SrvInfo, "ClientUid:", c1.GetUid())
	}
	defer c1.Close()

	c2 := NewClient(clientOption)
	SrvInfo, err = c2.Connect("vsoa", *our_addr)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("SrvInfo:", SrvInfo, "ClientUid:", c2.GetUid())
	}
	defer c2.Close()

	c3 := NewClient(clientOption)
	SrvInfo, err = c3.Connect("vsoa", *our_addr)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("SrvInfo:", SrvInfo, "ClientUid:", c3.GetUid())
	}
	defer c3.Close()

	req := protocol.NewMessage()

	req.Param, _ = json.RawMessage(`{"Test Num":1}`).MarshalJSON()
	reply, err := c1.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		DstParam := new(RpcTestParam)
		json.Unmarshal(reply.Param, DstParam)
		t.Log("Seq:", reply.SeqNo(), "Param:", DstParam, "Unmarshaled data:", DstParam.Num)
	}

	req.Param, _ = json.RawMessage(`{"Test Num":2}`).MarshalJSON()
	reply, err = c2.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		DstParam := new(RpcTestParam)
		json.Unmarshal(reply.Param, DstParam)
		t.Log("Seq:", reply.SeqNo(), "Param:", DstParam, "Unmarshaled data:", DstParam.Num)
	}

	req.Param, _ = json.RawMessage(`{"Test Num":3}`).MarshalJSON()
	reply, err = c3.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		DstParam := new(RpcTestParam)
		json.Unmarshal(reply.Param, DstParam)
		t.Log("Seq:", reply.SeqNo(), "Param:", DstParam, "Unmarshaled data:", DstParam.Num)
	}
}
