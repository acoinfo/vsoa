package client

import (
	"encoding/json"
	"flag"
	"go-vsoa/protocol"
	"testing"
)

var (
	datagram_addr = flag.String("datagram_addr", "localhost:3002", "server address")
)

type DatagramTestParam struct {
	Num int `json:"Test Num"`
}

func TestDatagram(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *datagram_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()

	req.Param, _ = json.RawMessage(`{"Test Num":123}`).MarshalJSON()

	_, err = c.Call("/datagram", protocol.TypeDatagram, protocol.ChannelNormal, req)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Datagram send done")
	}
}

func TestDatagramQuick(t *testing.T) {
	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *datagram_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()

	req.Param, _ = json.RawMessage(`{"Test Num":123}`).MarshalJSON()

	_, err = c.Call("/datagramQuick", protocol.TypeDatagram, protocol.ChannelQuick, req)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Datagram send done")
	}
}
