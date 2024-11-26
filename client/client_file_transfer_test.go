package client

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"testing"
	"time"

	"github.com/go-sylixos/go-vsoa/protocol"
	"github.com/go-sylixos/go-vsoa/server"
)

type FileTransferTestParam struct {
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
}

var (
	file_transfer_addr = flag.String("file_transfer_addr", "localhost:3006", "file_transfer server address")
)

//go:embed vsoa.png
var orginalFile []byte

func TestFileTransfer(t *testing.T) {
	startFileStreamServer(t)
	flag.Parse()

	// Do this to make sure the server is ready on slow machine
	time.Sleep(50 * time.Millisecond)

	var StreamTunID uint16

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *file_transfer_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()
	DstParam := new(FileTransferTestParam)

	reply, err := c.Call("/download", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		StreamTunID = reply.TunID()
		t.Log("Seq:", reply.SeqNo(), "Stream TunID:", StreamTunID)

		json.Unmarshal(reply.Param, DstParam)
		t.Log("FileName:", DstParam.FileName, "FileSize:", DstParam.FileSize)
	}

	receiveBuf := bytes.NewBufferString("")

	streamDone := make(chan error)

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
						streamDone <- err
						return
					} else {
						streamDone <- err
						return
					}
				}
				receiveBuf.Write(buf[:n])

				remainingSize := DstParam.FileSize - receiveBuf.Len()

				if remainingSize <= 0 {
					break
				}
			}

			if md5.Sum(orginalFile) != md5.Sum(receiveBuf.Bytes()) {
				streamDone <- errors.New("Stream file md5 not match!")
				return
			}

			t.Log("stream download file done successfully!")
			streamDone <- nil
		}()
	}

	d := <-streamDone
	cs.conn.Close()

	if d != nil {
		t.Fatal(d)
	}
}

func TestFileTransferIoCopy(t *testing.T) {
	startFileStreamServer(t)
	flag.Parse()

	// Do this to make sure the server is ready on slow machine
	time.Sleep(50 * time.Millisecond)

	var StreamTunID uint16

	clientOption := Option{
		Password: "123456",
	}

	c := NewClient(clientOption)
	_, err := c.Connect("vsoa", *file_transfer_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	req := protocol.NewMessage()
	DstParam := new(FileTransferTestParam)

	reply, err := c.Call("/download", protocol.TypeRPC, protocol.RpcMethodGet, req)
	if err != nil {
		if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
			t.Log("Pass: Invalid URL")
		} else {
			t.Fatal(err)
		}
	} else {
		StreamTunID = reply.TunID()
		t.Log("Seq:", reply.SeqNo(), "Stream TunID:", StreamTunID)

		json.Unmarshal(reply.Param, DstParam)
		t.Log("FileName:", DstParam.FileName, "FileSize:", DstParam.FileSize)
	}

	receiveBuf := bytes.NewBufferString("")

	streamDone := make(chan error)

	cs, err := c.NewClientStream(StreamTunID)
	if err != nil {
		t.Fatal(err)
	} else {
		go func() {
			io.CopyN(receiveBuf, cs, int64(DstParam.FileSize))

			if md5.Sum(orginalFile) != md5.Sum(receiveBuf.Bytes()) {
				streamDone <- errors.New("Stream file md5 not match!")
				return
			}

			t.Log("stream download file done successfully!")
			streamDone <- nil
		}()
	}

	d := <-streamDone
	cs.conn.Close()

	if d != nil {
		t.Fatal(d)
	}
}

func startFileStreamServer(t *testing.T) {
	// Init golang server
	serverOption := server.Option{
		Password: "123456",
	}
	s := server.NewServer("golang VSOA stream server", serverOption)

	// Register URL
	h := func(req, res *protocol.Message) {
		p := new(FileTransferTestParam)
		p.FileName = "vsoa_after.png"
		p.FileSize = len(orginalFile)

		var err error

		res.Param, err = json.Marshal(p)
		if err != nil {
			t.Fatal(err)
			return
		}

		ss, _ := s.NewServerStream(res)
		pushBuf := bytes.NewBuffer(orginalFile)
		receiveBuf := bytes.NewBufferString("")
		go func() {
			ss.ServeListener(pushBuf, receiveBuf)
			//t.Log("stream server receiveBuf:", receiveBuf.String())
		}()
	}
	s.On("/download", protocol.RpcMethodGet, h)

	go func() {
		_ = s.Serve("127.0.0.1:3006")
	}()
	//defer s.Close()
	// Done init golang stream server
}
