package server

import (
	"errors"
	"net"
	"testing"
	"time"
)

func TestServeQuickListenerReturnsErrServerClosedOnClosedSocket(t *testing.T) {
	s := NewServer("test", Option{})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()

	s.ln = ln
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.serveQuickListener("")
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		if s.qln != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("quick listener did not start in time")
		}
		time.Sleep(10 * time.Millisecond)
	}

	s.isShutdown.Store(true)
	if err := s.qln.Close(); err != nil {
		t.Fatalf("close udp listener: %v", err)
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, ErrServerClosed) {
			t.Fatalf("expected ErrServerClosed, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("quick listener did not exit after udp socket close")
	}
}

func TestCloseClosesDoneChan(t *testing.T) {
	s := NewServer("test", Option{})

	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer tcpLn.Close()

	udpLn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer udpLn.Close()

	s.ln = tcpLn
	s.qln = udpLn
	s.doneChan = make(chan struct{})
	s.isStarted.Store(true)

	doneChan := s.doneChan
	if err := s.Close(); err != nil {
		t.Fatalf("close server: %v", err)
	}

	select {
	case <-doneChan:
	case <-time.After(time.Second):
		t.Fatal("doneChan was not closed by Close")
	}
}
