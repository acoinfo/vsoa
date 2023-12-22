package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

// Client represents a VSOA client. (For NOW it's only RPC&ServInfo)
type ClientStream struct {
	addr  string
	tunid uint16

	conn net.Conn
}

func (client *Client) NewClientStream(tunid uint16) (cs *ClientStream, err error) {
	t := strconv.Itoa(int(tunid))
	n := 0
	if client != nil {
		n = strings.LastIndexByte(client.addr, ':')
	} else {
		return nil, fmt.Errorf("nil client")
	}

	address := client.addr[:n+1] + t

	var conn net.Conn

	conn, err = net.DialTimeout("tcp", address, client.option.ConnectTimeout)

	if err != nil {
		log.Printf("failed to dial server in stream mode: %v", err)
		return nil, err
	}

	return &ClientStream{
		addr:  address,
		tunid: tunid,
		conn:  conn,
	}, nil
}

func (cs *ClientStream) StopClientStream() (err error) {
	return cs.conn.Close()
}

func (cs *ClientStream) Read(buf []byte) (int, error) {
	return cs.conn.Read(buf)
}

func (cs *ClientStream) Write(writeBuf *bytes.Buffer) {
	io.CopyN(cs.conn, writeBuf, int64(writeBuf.Len()))
}
