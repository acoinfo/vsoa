// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: message.go Vehicle SOA protocal package.
//
// Author: Cheng.siyuan <chengsiyuan@acoinfo.com>

package protocol

import (
	"encoding/binary"
	"encoding/json"
	"net"
)

type ServInfoReqParam struct {
	Password     string `json:"passwd,omitempty"`
	PingInterval int    `json:"pingInterval,omitempty"`
	PingTimeout  int    `json:"pingTimeout,omitempty"`
	PingLost     int    `json:"pingLost,omitempty"`
}

type ServInfoResParam struct {
	Info string `json:"info"` // it should be JSON, but in real world it's a string
}

type ServInfoResData struct {
	ClientUid uint32 // this is used for Client to now it's Uid for the point server
}

const (
	ServInfoResAsString = iota
	ServInfoResAsJSON
)

func (q ServInfoReqParam) NewMessage(req *Message, laddr string) {
	req.SetMessageType(TypeServInfo)
	req.SetStatusType(StatusSuccess)
	// Add real quick channel id to fill the Tunid
	udpaddr, _ := net.ResolveUDPAddr("udp", laddr)
	req.SetTunId(udpaddr.AddrPort().Port())
	req.SetValidTunid()
	req.SetReply(false)

	// set everthing inside ServInfoReqParam, It can be empty
	req.Param, _ = json.Marshal(q)
	req.Data = nil
}

func (s ServInfoResParam) NewMessage(ResType int, res *Message, ClientUid uint32) {
	res.SetMessageType(TypeServInfo)
	res.SetStatusType(StatusSuccess)
	res.SetReply(true)

	if cap(res.Data) >= 4 { // reuse Data
		res.Data = res.Data[:4]
	} else {
		res.Data = make([]byte, 4)
	}

	switch ResType {
	case ServInfoResAsJSON:
		res.Param, _ = json.Marshal(s)
		binary.BigEndian.PutUint32(res.Data, uint32(ClientUid))
	case ServInfoResAsString:
		fallthrough // it still be a string
	default:
		res.Param = json.RawMessage(s.Info)
		binary.BigEndian.PutUint32(res.Data, uint32(ClientUid))
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

func GetClientUid(u []byte) uint32 {
	return binary.BigEndian.Uint32(u)
}
