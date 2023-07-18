package client

import (
	"encoding/json"
	"flag"
	"go-vsoa/protocol"
	"go-vsoa/server"
	"testing"
)

var (
	rpc_addr = flag.String("rpc_addr", "localhost:3003", "server address")
)

type RpcTestParam struct {
	Num int `json:"Test Num"`
}

func TestRPC(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *rpc_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()
	reply := protocol.NewMessage()

	reply, err = c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		t.Log("Seq:", reply.SeqNo(), "Param:", (reply.Param))
	}

	req.Param, _ = json.RawMessage(`{"Test Num":123}`).MarshalJSON()

	reply, err = c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
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

func init() {
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
