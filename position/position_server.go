package position

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"runtime"
)

func (pl *PositionList) ServePositionListener(address net.UDPAddr) (err error) {
	pln, err := net.ListenUDP("udp", &address)
	if err != nil {
		log.Fatal(err)
	}
	defer pln.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := pln.ReadFromUDP(buf)
		//_, _, err = pln.ReadFromUDP(buf)
		if err != nil {
			continue
		} else {
			r := bytes.NewBuffer(buf)
			p, err := pl.processLoopUpRequest(r, n)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("Position server has closed this connection: %s", pln.LocalAddr().String())
				}
				log.Println(err)
				continue
			}
			outBuf := p.prepareLookUpResponse()
			if outBuf != nil {
				pln.WriteToUDP(outBuf, addr)
			}
		}
	}
}

// Decode decodes a message from reader.
func (pl *PositionList) processLoopUpRequest(r io.Reader, n int) (p *Position, err error) {
	defer func() {
		if err := recover(); err != nil {
			var errStack = make([]byte, 1024)
			n := runtime.Stack(errStack, true)
			log.Printf("panic in message decode: %v, stack: %s", err, errStack[:n])
		}
	}()

	buffer := make([]byte, 1024)

	_, err = io.ReadFull(r, buffer[:n])
	if err != nil {
		return nil, err
	}

	var Param LookUpRequest
	json.Unmarshal(buffer[:n], &Param)

	p = pl.lookUp(Param.Name)
	if p == nil {
		return nil, errors.New("Position: server " + Param.Name + " not found")
	}

	return p, nil
}

func (p Position) prepareLookUpResponse() []byte {
	if net.ParseIP(p.IP) == nil {
		return nil
	}
	if p.Port == 0 {
		return nil
	}

	buffer, err := json.Marshal(p)
	if err != nil {
		return nil
	}

	return buffer
}
