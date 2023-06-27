// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: message.go Vehicle SOA protocal package.
//
// Author: Cheng.siyuan <chengsiyuan@acoinfo.com>

package protocol

import "encoding/json"

type ServInfoReqParam struct {
	Password     string `json:"passwd,omitempty"`
	PingInterval int    `json:"pingInterval,omitempty"`
	PingTimeout  int    `json:"pingTimeout,omitempty"`
	PingLost     int    `json:"pingLost,omitempty"`
}

type ServInfoResParam struct {
	Info string `json:"info"` // it should be JSON, but in real world it's a string
}

const (
	ServInfoResAsString = iota
	ServInfoResAsJSON
)

func (q ServInfoReqParam) NewMessage(req *Message) {
	req.SetMessageType(TypeServInfo)
	req.SetStatusType(StatusSuccess)
	req.SetValidTunid()
	req.SetReply(false)

	// set everthing inside ServInfoReqParam, It can be empty
	req.Param, _ = json.Marshal(q)
	req.Data = nil
}

func (s ServInfoResParam) NewMessage(ResType int, res *Message) {
	res.SetMessageType(TypeServInfo)
	res.SetStatusType(StatusSuccess)
	res.SetReply(true)

	switch ResType {
	case ServInfoResAsJSON:
		res.Param, _ = json.Marshal(s)
	case ServInfoResAsString:
		fallthrough // it still be a string
	default:
		res.Param = json.RawMessage(s.Info)
	}
}

func DecodeServInfo(m json.RawMessage) string {
	infoParam := new(ServInfoResParam)
	err := json.Unmarshal([]byte(m), infoParam)
	if err != nil {
		return string(m)
	} else {
		return infoParam.Info
	}
}
