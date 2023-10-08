package position

import (
	"flag"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	position_addr     = flag.String("position_addr", "localhost:6002", "position address")
	StartPositionOnce sync.Once
)

func TestPositionLookUP(t *testing.T) {
	StartPositionOnce.Do(func() {
		positionServerStart()
	})
	flag.Parse()

	p := new(Position)
	err := p.LookUp("light_server", *position_addr, 500*time.Millisecond)
	if err == ErrLookUpTimeOut || err == ErrServerNotFound {
		t.Fatal(err)
	} else if err != nil {
		t.Fatal(err)
	}

	t.Log(p)
}

func TestPositionLookUPTimeout(t *testing.T) {
	StartPositionOnce.Do(func() {
		positionServerStart()
	})
	flag.Parse()

	p := new(Position)
	err := p.LookUp("light_server__", *position_addr, 500*time.Millisecond)
	if err == ErrLookUpTimeOut || err == ErrServerNotFound {
		t.Log(err)
	} else if err != nil {
		t.Fatal(err)
	}
}

func positionServerStart() {
	pl := NewPositionList()
	pl.Add(*NewPosition("light_server", 1, "127.0.0.1", 6001, false))

	go pl.ServePositionListener(net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 6002,
	})
}
