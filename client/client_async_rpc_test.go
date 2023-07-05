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
	_, err := c.Connect("vsoa", *rpc_async_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req1 := protocol.NewMessage()
	req2 := protocol.NewMessage()
	reply := protocol.NewMessage()

	biddata := &RpcAsyncTestParam{
		// 255KB Param
		BigData: string(*makeLargeByteArray('A', 255)),
	}
	req1.Param, _ = json.Marshal(biddata)
	biddata = &RpcAsyncTestParam{
		// 255KB Param
		BigData: string(*makeLargeByteArray('B', 255)),
	}
	req2.Param, _ = json.Marshal(biddata)

	// Actually We don't need to care Call1 using seq:1 or not, since it's async call
	Call1 := c.Go("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req1, reply, nil).Done
	Call2 := c.Go("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req2, reply, nil).Done

	for i := 0; i < 2; i++ {
		select {
		case call := <-Call1:
			t.Log("Call1 Data should be like A, Param LEN:", len(call.Reply.Param))
			logAsyncCall(call, t)
		case call := <-Call2:
			t.Log("Call2 Data should be like B, Param LEN:", len(call.Reply.Param))
			logAsyncCall(call, t)
		}
	}
}

func TestRPCMixed(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *rpc_async_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	reqSync := protocol.NewMessage()
	req1 := protocol.NewMessage()
	req2 := protocol.NewMessage()
	reply := protocol.NewMessage()

	biddata := &RpcAsyncTestParam{
		// Larger than 256KB Message Test
		BigData: string(*makeLargeByteArray('A', 256)),
	}
	req1.Param, _ = json.Marshal(biddata)
	biddata = &RpcAsyncTestParam{
		// Larger than 256KB Message Test
		BigData: string(*makeLargeByteArray('B', 256)),
	}
	req2.Param, _ = json.Marshal(biddata)

	reply, err = c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, reqSync)
	if err != nil {
		t.Fatal(err)
	} else {
		if reply.SeqNo() != 1 {
			t.Fatal("Not sync")
		}
	}

	// Actually We don't need to care Call1 using seq:1 or not, since it's async call
	Call1 := c.Go("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req1, reply, nil).Done
	Call2 := c.Go("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req2, reply, nil).Done

	reply, err = c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, reqSync)
	if err != nil {
		t.Fatal(err)
	} else {
		if reply.SeqNo() != 2 {
			t.Fatal("Not sync")
		}
	}

	for i := 0; i < 2; i++ {
		select {
		case replyCall := <-Call1:
			if replyCall.Error == protocol.ErrMessageTooLong {
				t.Logf("passed ErrMessageTooLong test")
			} else {
				t.Fatalf("failed to test ErrMessageTooLong real err  like: %v", replyCall.Error)
			}
		case replyCall := <-Call2:
			if replyCall.Error == protocol.ErrMessageTooLong {
				t.Logf("passed ErrMessageTooLong test")
			} else {
				t.Fatalf("failed to test ErrMessageTooLong real err  like: %v", replyCall.Error)
			}
		}
	}

	reply, err = c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, reqSync)
	if err != nil {
		t.Fatal(err)
	} else {
		if reply.SeqNo() != 5 {
			t.Fatal("Not sync")
		}
	}
}

func makeLargeByteArray(raw byte, KB int) *[]byte {
	var _KB int = 1024 * KB
	buffer := make([]byte, _KB)
	tmp := make([]byte, 1)
	tmp[0] = raw

	for i := 0; i < _KB; i++ {
		copy(buffer[i:], tmp)
	}

	return &buffer
}

func logAsyncCall(call *Call, t *testing.T) {
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
