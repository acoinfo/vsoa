package position

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"time"
)

type LookUpRequest struct {
	Name string `json:"name"`
}

// ErrShutdown connection is closed.
var (
	ErrLookUpTimeOut  = errors.New("Position: LookUp timeout")
	ErrServerNotFound = errors.New("Position: server not found")
)

func (p *Position) LookUp(name string, position_addr string, timeout time.Duration) (err error) {
	var qconn *net.UDPConn
	var saddr *net.UDPAddr

	saddr, err = net.ResolveUDPAddr("udp", position_addr)
	if err != nil {
		log.Printf("failed to resolve UDP address: %v", err)
		return err
	}

	qconn, err = net.DialUDP("udp", nil, saddr)
	if err != nil {
		log.Printf("failed to dial position server: %v", err)
		return err
	}
	defer qconn.Close()

	l := new(LookUpRequest)
	l.Name = name
	buffer := prepareLookUpRequest(*l)

	_, err = qconn.Write(buffer)
	if err != nil {
		return err
	}

	done := make(chan *Position, 1)

	go func() {
		pbuffer := make([]byte, 1024)
		n, addr, err := qconn.ReadFromUDP(pbuffer)
		if err != nil {
			p = nil
		} else if addr.String() != saddr.String() {
			p = nil
		}

		err = json.Unmarshal(pbuffer[:n], &p)
		if err != nil {
			p = nil
		}

		done <- p
	}()

	select {
	case <-done:
		qconn.Close()
		if p != nil {
			return nil
		} else {
			return ErrServerNotFound
		}

	case <-time.After(timeout):
		qconn.Close()
		return ErrLookUpTimeOut
	}
}

func prepareLookUpRequest(l LookUpRequest) []byte {
	buffer, err := json.Marshal(l)
	if err != nil {
		return nil
	}

	return buffer
}
