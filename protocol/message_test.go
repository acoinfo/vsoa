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
	"encoding/json"
	"testing"
)

type TestParam struct {
	Password string `json:"passwd"`
}

func TestMessage(t *testing.T) {
	req := NewMessage()
	req.SetMessageType(TypeServInfo)
	req.SetStatusType(StatusSuccess)
	req.SetReply(false)

	req.SetSeqNo(1234567890)

	req.URL = []byte("vsoa://test-message/v1")
	// Use json.RawMessage to fill Param;
	req.Param, _ = json.RawMessage(`{"passwd":"123456"}`).MarshalJSON()
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

	res.Decode(&buf)

	if res.Version() != version {
		t.Errorf("expect %d but got %d", version, res.Version())
	}

	if res.SeqNo() != 1234567890 {
		t.Errorf("expect 1234567890 but got %d", res.SeqNo())
	}

	if string(res.URL) != "vsoa://test-message/v1" {
		t.Errorf("got wrong URL: %v", res.URL)
	}

	DstParam := new(TestParam)
	err = json.Unmarshal(res.Param, DstParam)
	if err != nil {
		t.Error("Unmarshal Param JSON err: ", err)
	}
	if DstParam.Password != "123456" {
		t.Errorf("got wrong Param: %v", DstParam.Password)
	}

	infostr := `{"name":"Golang VSOA server"}`
	infoParam := new(ServInfoResParam)
	err = json.Unmarshal([]byte(infostr), infoParam)
	if err != nil {
		t.Log("Unmarshal Param JSON err: ", err)
	}

	if hex.EncodeToString(res.Data) != hex.EncodeToString([]byte{
		0xff, 0xee, 0xdd,
	}) {
		t.Errorf("got wrong payload: %v", string(res.Data))
	}
}
