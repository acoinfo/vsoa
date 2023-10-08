package client

import (
	"bytes"
	"flag"
	"go-vsoa/protocol"
	"go-vsoa/server"
	"io"
	"testing"
	"time"
)

var (
	stream_addr = flag.String("stream_addr", "localhost:3002", "stream server address")
)

func TestStream(t *testing.T) {
	startStreamServer(t)
	flag.Parse()

	var StreamTunID uint16

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *stream_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()

	reply, err := c.Call("/read", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		StreamTunID = reply.TunID()
		t.Log("Seq:", reply.SeqNo(), "Stream TunID:", StreamTunID)
	}

	receiveBuf := bytes.NewBufferString("")

	cs, err := c.NewClientStream(StreamTunID)
	if err != nil {
		t.Fatal(err)
	} else {
		go func() {
			buf := make([]byte, 32*1024)
			for {
				n, err := cs.Read(buf)
				if err != nil {
					// EOF means stream cloesed
					if err == io.EOF {
						break
					} else {
						t.Log(err)
						break
					}
				}
				receiveBuf.Write(buf[:n])
				t.Log("stream receiveBuf:", receiveBuf.String())

				// Push data back to stream server
				cs.Write(receiveBuf.Bytes())

				//In this test, we just receive little data from server, so we just stop here
				goto STOP
			}

		STOP:
			cs.conn.Close()
		}()
	}

	// don't close too quick before server handle the Call
	time.Sleep(5 * time.Millisecond)
}

func startStreamServer(t *testing.T) {
	// Init golang server
	serverOption := server.Option{
		Password: "123456",
	}
	s := server.NewServer("golang VSOA stream server", serverOption)

	// Register URL
	h := func(req, res *protocol.Message) {
		ss, _ := s.NewSeverStream(res)
		pushBuf := bytes.NewBufferString("12345678909876543212345678910")
		receiveBuf := bytes.NewBufferString("")
		go func() {
			ss.ServeListener(pushBuf, receiveBuf)
			t.Log("stream server receiveBuf:", receiveBuf.String())
		}()
	}
	s.AddRpcHandler("/read", protocol.RpcMethodGet, h)

	go func() {
		_ = s.Serve("127.0.0.1:3002")
	}()
	//defer s.Close()
	// Done init golang stream server
}
