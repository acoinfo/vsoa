package client

import (
	"strings"
	"time"

	"github.com/acoinfo/vsoa/protocol"
)

type regulator struct {
	interval time.Duration
	stop     chan int
}

type clientSlot struct {
	hasData bool
	raw     *protocol.Message
	handler func(*protocol.Message)
}

func (client *Client) NoopPublish(m *protocol.Message) {}

func (client *Client) Slot(URL string, onPublish func(m *protocol.Message)) error {
	var err error = nil

	if client.slotList == nil {
		client.slotList = make(map[string](*clientSlot))
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	cs := new(clientSlot)
	cs.hasData = false
	if onPublish != nil {
		cs.handler = onPublish
	} else {
		cs.handler = client.NoopPublish
	}

	client.slotList[URL] = cs
	return err
}

// UnSubscribe server URL;
// free callback to the URL with father URL
func (client *Client) UnSlot(URL string) error {
	var err error = nil
	if !client.IsAuthed() {
		err = ErrUnAuthed
		return err
	}

	if client.slotList == nil {
		// Already unSubscribe
		return nil
	}

	if _, ok := client.slotList[URL]; ok {
		client.mutex.Lock()
		delete(client.slotList, URL)
		client.mutex.Unlock()
	} else if strings.HasSuffix(URL, "/") {
		if client.slotList[URL[:len(URL)-1]] != nil {
			client.mutex.Lock()
			delete(client.slotList, URL[:len(URL)-1])
			client.mutex.Unlock()
		}
		// Already unSubscribe
		return nil
	}

	return err
}

func (client *Client) StartRegulator(interval time.Duration) error {
	if interval < 1*time.Millisecond {
		return ErrRegulatorTooFast
	}

	client.regulator.interval = interval
	ticker := time.NewTicker(interval)

	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.hasRegulator {
		return ErrStartRegulator
	}
	client.hasRegulator = true

	client.regulator.stop = make(chan int)

	go func() {
		defer ticker.Stop()
		defer close(client.regulator.stop)
		for {
			select {
			case <-ticker.C:
				client.regulatorHandler()
			case <-client.regulator.stop:
				return
			}
		}
	}()

	return nil
}

func (client *Client) StopRegulator() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if !client.hasRegulator {
		return ErrStopRegulator
	}

	client.regulator.stop <- 1
	client.hasRegulator = false

	return nil
}

func (client *Client) regulatorUpdator(res *protocol.Message) {
	if !client.hasRegulator {
		return
	}

	if client.slotList == nil {
		return
	}
	if _, ok := client.slotList[string(res.URL)]; ok {
		client.slotList[string(res.URL)].hasData = true
		client.slotList[string(res.URL)].raw = res
	}
}

func (client *Client) regulatorHandler() {
	if client.slotList == nil {
		return
	}
	for _, v := range client.slotList {
		if v.hasData {
			v.hasData = false
			v.handler(v.raw)
		}
	}
}
