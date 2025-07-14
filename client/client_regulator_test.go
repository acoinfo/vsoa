package client

import (
	"flag"
	"testing"
	"time"

	"github.com/acoinfo/go-vsoa/protocol"
)

// TestSlot is a test function that tests the functionality of the Slot method in the Client struct.
//
// It starts the server and waits for it to be ready.
// Then it creates a callback object and sets the T field to the provided testing.T object.
// It parses the command line flags. It creates a new client with a password option.
// It starts the regulator with a 2-second interval.
func TestSlot(t *testing.T) {
	startServer()

	// Do this to make sure the server is ready on slow machine
	time.Sleep(50 * time.Millisecond)

	cb := new(callback)
	cb.T = t

	flag.Parse()

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	c.StartRegulator(2 * time.Second)

	_, err := c.Connect("vsoa", *publish_server_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// client don't know if it's quick channel or not
	err = c.Subscribe("/p", c.NoopPublish)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	}

	c.Slot("/p", cb.getPublishParam)

	time.Sleep(3 * time.Second)

	c.StopRegulator()

	time.Sleep(4 * time.Second)

	c.StartRegulator(2 * time.Second)
	time.Sleep(3 * time.Second)
}
