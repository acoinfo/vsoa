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
	req3 := protocol.NewMessage()
	req4 := protocol.NewMessage()
	req5 := protocol.NewMessage()
	reply := protocol.NewMessage()

	req1.URL = []byte("/a/b/c")
	req2.URL = []byte("/a/b/c")
	req3.URL = []byte("/a/b/c")
	req4.URL = []byte("/a/b/c")
	req5.URL = []byte("/a/b/c")

	biddata := &RpcAsyncTestParam{
		BigData: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	req1.Param, _ = json.Marshal(biddata)
	biddata = &RpcAsyncTestParam{
		BigData: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	}
	req2.Param, _ = json.Marshal(biddata)

	// TODO: we expact to send Call1 with Data "AAA..A" Call2 to send Data "BBB..B".
	// But it send unordered.
	// The Reason is inside Go func client sync the seq but async fill in reqs
	Call1 := c.Go(protocol.TypeRPC, protocol.RpcMethodGet, req1, reply, nil).Done
	Call2 := c.Go(protocol.TypeRPC, protocol.RpcMethodGet, req2, reply, nil).Done

	// Even data in Call be unordered. NodeJS server send the right data back to client, A for A, B for B.
	select {
	case call := <-Call1:
		err = call.Error
		reply = call.Reply
		SrcParam := new(RpcAsyncTestParam)
		json.Unmarshal(*call.Param, SrcParam)
		DstParam := new(RpcAsyncTestParam)
		json.Unmarshal(reply.Param, DstParam)
		if SrcParam.BigData == DstParam.BigData {
			t.Log("Reply seq No:", reply.SeqNo(), "Data like: ", DstParam.BigData[:1])
		} else {
			t.Fatal("error Date miss match")
		}
	}
	select {
	case call := <-Call2:
		err = call.Error
		reply = call.Reply
		SrcParam := new(RpcAsyncTestParam)
		json.Unmarshal(*call.Param, SrcParam)
		DstParam := new(RpcAsyncTestParam)
		json.Unmarshal(reply.Param, DstParam)
		if SrcParam.BigData == DstParam.BigData {
			t.Log("Reply seq No:", reply.SeqNo(), "Data like: ", DstParam.BigData[:1])
		} else {
			t.Fatal("error Date miss match")
		}
	}
}
