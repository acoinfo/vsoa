package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"go-vsoa/protocol"
	"log"
	"net"
)

// Connect connects the server via specified network.
// ServInfo Shack hand is needed cause VSOA protocol
func (client *Client) Connect(network, address string) (ServerInfo string, err error) {
	var conn net.Conn
	var qconn *net.UDPConn

	switch network {
	default:
		conn, err = newDirectConn(client, address)

		if err == nil && conn != nil {
			client.Conn = conn
			client.r = bufio.NewReaderSize(conn, ReaderBuffsize)
			// c.w = bufio.NewWriterSize(conn, WriterBuffsize)

			// start reading and writing since connected
			go client.input()
		}

		qconn, err = newQuickConn(client, address)
		{
			if err == nil && conn != nil {
				client.QConn = qconn
				// TODO: add Quick channel input thread since connected
			}
		}
	}

	req := protocol.NewMessage()
	reply := protocol.NewMessage()

	reply, err = client.Call("", protocol.TypeServInfo, protocol.RpcMethodGet, req)
	if err != nil {
		return "", err
	}

	client.uid = protocol.GetClientUid(reply.Data)
	return protocol.DecodeServInfo(reply.Param), err
}

func newQuickConn(c *Client, address string) (*net.UDPConn, error) {
	var qconn *net.UDPConn
	var saddr *net.UDPAddr
	var err error

	if c == nil {
		err = fmt.Errorf("nil client")
		return nil, err
	}

	saddr, err = net.ResolveUDPAddr("udp", address)

	qconn, err = net.DialUDP("udp", nil, saddr)
	if err != nil {
		log.Printf("failed to dial server quick path: %v", err)
		return nil, err
	}

	return qconn, nil
}

func newDirectConn(c *Client, address string) (net.Conn, error) {
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
		tlsConn, err = tls.DialWithDialer(dialer, "tcp", address, c.option.TLSConfig)
		// or conn:= tls.Client(netConn, &config)
		conn = net.Conn(tlsConn)
	} else {
		conn, err = net.DialTimeout("tcp", address, c.option.ConnectTimeout)
	}

	if err != nil {
		log.Printf("failed to dial server: %v", err)
		return nil, err
	}

	return conn, nil
}
