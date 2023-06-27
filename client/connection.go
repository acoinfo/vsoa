package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

// Connect connects the server via specified network.
func (client *Client) Connect(network, address string) error {
	var conn net.Conn
	var err error

	switch network {

	default:
		conn, err = newDirectConn(client, network, address)

		if err == nil && conn != nil {
			client.Conn = conn
			client.r = bufio.NewReaderSize(conn, ReaderBuffsize)
			// c.w = bufio.NewWriterSize(conn, WriterBuffsize)

			// start reading and writing since connected
			go client.input()
		}
	}

	return err
}

func newDirectConn(c *Client, network, address string) (net.Conn, error) {
	var conn net.Conn
	var tlsConn *tls.Conn
	var err error

	if c == nil {
		err = fmt.Errorf("nil client")
		return nil, err
	}

	if c.option.TLSConfig != nil {
		dialer := &net.Dialer{
			Timeout: c.option.ConnectTimeout,
		}
		tlsConn, err = tls.DialWithDialer(dialer, network, address, c.option.TLSConfig)
		// or conn:= tls.Client(netConn, &config)
		conn = net.Conn(tlsConn)
	} else {
		conn, err = net.DialTimeout(network, address, c.option.ConnectTimeout)
	}

	if err != nil {
		log.Printf("failed to dial server: %v", err)
		return nil, err
	}

	return conn, nil
}
