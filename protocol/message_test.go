// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: message_test.go Vehicle SOA protocal package.
//
// Author: Cheng.siyuan <chengsiyuan@acoinfo.com>

package protocol

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestMessage(t *testing.T) {
	req := NewMessage()
	req.SetMessageType(TypeServInfo)
	req.SetStatusType(StatusSuccess)
	req.SetReply(false)

	req.SetSeqNo(1234567890)

	req.URL = []byte("vsoa://test-message/v1")
	req.Param = []byte("{\"password\":\"123456\"}")
	req.Data = []byte{
		0xff, 0xee, 0xdd,
	}

	var buf bytes.Buffer
	_, err := req.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	res.SetReply(true)

	if res.Version() != version {
		t.Errorf("expect %d but got %d", version, res.Version())
	}

	if res.SeqNo() != 1234567890 {
		t.Errorf("expect 1234567890 but got %d", res.SeqNo())
	}

	if string(res.URL) != "vsoa://test-message/v1" {
		t.Errorf("got wrong URL: %v", res.URL)
	}

	if string(res.Param) != "{\"password\":\"123456\"}" {
		t.Errorf("got wrong Param: %v", res.Param)
	}

	if hex.EncodeToString(res.Data) != hex.EncodeToString([]byte{
		0xff, 0xee, 0xdd,
	}) {
		t.Errorf("got wrong payload: %v", string(res.Data))
	}
}
