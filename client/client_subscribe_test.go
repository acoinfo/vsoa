package client

import (
	"encoding/json"
	"flag"
	"go-vsoa/protocol"
	"testing"
	"time"
)

var (
	publish_server_addr = flag.String("publish_server_addr", "localhost:3002", "server address")
)

type PublishTestParam struct {
	Publish string `json:"publish"`
}

type callback struct {
	T *testing.T
}

func TestSub(t *testing.T) {
	cb := new(callback)
	cb.T = t

	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *publish_server_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// client don't know if it's quick channel or not
	c.Subscribe("/p", cb.getPublishParam)

	time.Sleep(5 * time.Second)

	c.UnSubscribe("/p")

	time.Sleep(5 * time.Second)

	c.Subscribe("/p", cb.getPublishParam)

	time.Sleep(5 * time.Second)
}

// User can create callback struct to put/get more info into callback func
func (c callback) getPublishParam(m *protocol.Message) {
	DstParam := new(PublishTestParam)
	json.Unmarshal(m.Param, DstParam)
	c.T.Log("Param:", DstParam.Publish)
}
