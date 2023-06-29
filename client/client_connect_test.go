package client

import (
	"flag"
	"go-vsoa/protocol"
	"testing"
)

var (
	addr = flag.String("addr", "localhost:3002", "server address")
)

func TestGoodConnect(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	SrvInfo, err := c.Connect("tcp", *addr)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("SrvInfo:", SrvInfo)
	}
	defer c.Close()
}

func TestWorngPasswdConnect(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "12346",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("tcp", *addr)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusPassword)) {
			t.Log("passed Passwd Err test")
		} else {
			t.Fatal(err)
		}
	}
	defer c.Close()
}
