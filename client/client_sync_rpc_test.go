package client

import (
	"encoding/json"
	"flag"
	"go-vsoa/protocol"
	"testing"
)

var (
	rpc_addr = flag.String("rpc_addr", "localhost:3003", "server address")
)

type RpcTestParam struct {
	Num int `json:"Test Num"`
}

// TestRPC is a test function that performs RPC calls.
//
// TestRPC sets up a server, parses flags, creates a client with a password, connects to a specified address,
// and makes RPC calls to "/a/b/c" with different parameters. It logs the sequence number and parameters of the reply.
// The function also handles errors and logs specific messages for invalid URLs.
func TestRPC(t *testing.T) {
	startServer()
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

	reply, err := c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
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
