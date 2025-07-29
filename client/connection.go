package client

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/acoinfo/vsoa/position"
	"github.com/acoinfo/vsoa/protocol"
)

var (
	Type_URL = "VSOA_URL"
)

func (client *Client) SetPosition(address string) (err error) {
	parts := strings.Split(address, ":")
	if net.ParseIP(parts[0]) == nil && parts[0] != "localhost" {
		return errors.New("invalid position address should be IPv4/IPv6 address")
	}
	client.position = address
	return nil
}

// Connect connects the server via specified network.
// ServInfo Shack hand is needed cause VSOA protocol
// TODO: add position logic.
func (client *Client) Connect(vsoa_or_VSOA_URL, address_or_URL string) (ServerInfo string, err error) {
	if !client.option.AutoReconnect {
		return client.connectOnce(vsoa_or_VSOA_URL, address_or_URL)
	}

	for {
		serverInfo, err := client.connectOnce(vsoa_or_VSOA_URL, address_or_URL)
		if err == nil {
			return serverInfo, nil
		}
		time.Sleep(client.option.ReconnectInterval)
	}
}

func (client *Client) connectOnce(vsoa_or_VSOA_URL, address_or_URL string) (ServerInfo string, err error) {
	var conn net.Conn
	var qconn *net.UDPConn

	client.addr = address_or_URL
	client.connType = vsoa_or_VSOA_URL

	// check client options is valid
	if client.option.PingTurbo != 0 {
		if client.option.PingTurbo < 25 || client.option.PingTurbo > 1000 {
			return "", errors.New("invalid PingTurbo value should be between 25 and 1000 or 0")
		} else if client.option.PingInterval == 0 {
			return "", errors.New("PingInterval must be set with PingTurbo")
		} else if (client.option.PingInterval*1000)%client.option.PingTurbo != 0 {
			return "", errors.New("PingInterval must be multiple of PingTurbo")
		}
	}

	switch vsoa_or_VSOA_URL {
	case "VSOA_URL":
		//TODO: If position server change address, need to update all relevant connections
		if client.position == "" {
			return "", errors.New("position server not set with client.SetPosition()")
		}
		p := new(position.Position)

		parts := strings.Split(address_or_URL, "://")
		if len(parts) != 2 {
			err := p.LookUp(address_or_URL, client.position, 500*time.Millisecond)
			if err != nil {
				return "", err
			}
		} else {
			parts := strings.Split(address_or_URL, "://")
			err := p.LookUp(parts[1], client.position, 500*time.Millisecond)
			if err != nil {
				return "", err
			}
		}

		client.addr = p.IP + ":" + strconv.Itoa(p.Port)
		println("client.addr", client.addr)
		fallthrough
	default:
		conn, err = newDirectConn(client, client.addr)

		if err == nil && conn != nil {
			client.Conn = conn
			client.r = bufio.NewReaderSize(conn, ReaderBuffsize)

			// start reading and writing since connected
			go client.input()
		} else {
			return "", err
		}

		qconn, err = newQuickConn(client, client.addr)
		{
			if err == nil && qconn != nil {
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

	if client.option.OnConnect != nil {
		go client.option.OnConnect(client)
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
		return nil, err
	}

	return conn, nil
}
