package protocol

import (
	"testing"
)

func TestRes(t *testing.T) {
	m := &ServInfoResParam{
		Info: "TestRes",
	}

	res := NewMessage()
	m.NewMessage(ServInfoResAsJSON, res, 0x10)

	if GetClientUid(res.Data) != 0x10 {
		t.Fatalf("Client Uid should be 0x10, we got %x", GetClientUid(res.Data))
	} else {
		t.Log("Pass ServInfo Res test")
	}
}
