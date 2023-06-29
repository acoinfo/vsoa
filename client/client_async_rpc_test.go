package client

import (
	"encoding/json"
	"flag"
	"go-vsoa/protocol"
	"testing"
)

var (
	rpc_async_addr = flag.String("rpc_async_addr", "localhost:3002", "server address")
)

type RpcAsyncTestParam struct {
	BigData string `json:"Test Big Data"`
}

func TestRPCAsync(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("tcp", *rpc_async_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req1 := protocol.NewMessage()
	req2 := protocol.NewMessage()
	reply := protocol.NewMessage()

	req1.URL = []byte("/a/b/c")
	req2.URL = []byte("/a/b/c")

	biddata := &RpcAsyncTestParam{
		BigData: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	req1.Param, _ = json.Marshal(biddata)
	biddata = &RpcAsyncTestParam{
		BigData: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	}
	req2.Param, _ = json.Marshal(biddata)

	// Actually We don't need to care Call1 using seq:1 or not, since it's async call
	Call1 := c.Go(protocol.TypeRPC, protocol.RpcMethodGet, req1, reply, nil).Done
	Call2 := c.Go(protocol.TypeRPC, protocol.RpcMethodGet, req2, reply, nil).Done

	for i := 0; i < 2; i++ {
		select {
		case call := <-Call1:
			t.Log("Call1 Data should be like A")
			logAsyncCall(call, t)
		case call := <-Call2:
			t.Log("Call2 Data should be like B")
			logAsyncCall(call, t)
		}
	}
}

func logAsyncCall(call *RpcCall, t *testing.T) {
	reply := call.Reply
	SrcParam := new(RpcAsyncTestParam)
	json.Unmarshal(*call.Param, SrcParam)
	DstParam := new(RpcAsyncTestParam)
	json.Unmarshal(reply.Param, DstParam)
	if SrcParam.BigData == DstParam.BigData {
		t.Log("Data like: ", DstParam.BigData[:1])
	} else {
		t.Fatal("error Date miss match")
	}
}
