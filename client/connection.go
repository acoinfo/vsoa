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
// TODO: add position logic.
func (client *Client) Connect(network, address string) (ServerInfo string, err error) {
	var conn net.Conn
	var qconn *net.UDPConn

	client.addr = address

	switch network {
	default:
		conn, err = newDirectConn(client, address)

		if err == nil && conn != nil {
			client.Conn = conn
			client.r = bufio.NewReaderSize(conn, ReaderBuffsize)

			// start reading and writing since connected
			go client.input()
		} else {
			return "", err
		}

		qconn, err = newQuickConn(client, address)
		{
			if err == nil && conn != nil {
				client.QConn = qconn
				client.qr = bufio.NewReaderSize(qconn, ReaderBuffsize)
				go client.qinput()
			} else {
				return "", err
			}
		}
	}

	req := protocol.NewMessage()

	reply, err := client.Call("", protocol.TypeServInfo, protocol.RpcMethodGet, req)
	if err != nil {
		return "", err
	}

	client.mutex.Lock()
	client.authed = true
	// this is used for Quick channel
	client.uid = protocol.GetClientUid(reply.Data)
	client.mutex.Unlock()

	if client.option.PingInterval != 0 {
		go client.pingLoop()
	}

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

	saddr, _ = net.ResolveUDPAddr("udp", address)

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

// reConnect connects the server via specified network.
// ServInfo Shack hand is needed cause VSOA protocol
// TODO: add position logic.
func (client *Client) reConnect(network string) (err error) {
	var conn net.Conn
	var qconn *net.UDPConn

	client.Close()
	client.clearClient()
	// Kill input/qinput/pingloop go func before start a new client

	switch network {
	default:
		conn, err = newDirectConn(client, client.addr)

		if err == nil && conn != nil {
			client.Conn = conn
			client.r = bufio.NewReaderSize(conn, ReaderBuffsize)

			// start reading and writing since connected
			go client.input()
		} else {
			return err
		}

		qconn, err = newQuickConn(client, client.addr)
		{
			if err == nil && conn != nil {
				client.QConn = qconn
				client.qr = bufio.NewReaderSize(qconn, ReaderBuffsize)
				go client.qinput()
			} else {
				return err
			}
		}
	}

	req := protocol.NewMessage()

	reply, err := client.Call("", protocol.TypeServInfo, protocol.RpcMethodGet, req)
	if err != nil {
		return err
	}

	client.mutex.Lock()
	client.authed = true
	// this is used for Quick channel
	client.uid = protocol.GetClientUid(reply.Data)
	client.mutex.Unlock()

	if client.option.PingInterval != 0 {
		go client.pingLoop()
	}

	return err
}
