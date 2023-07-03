// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: type.go Vehicle SOA protocal package.
//
// Author: Cheng.siyuan <chengsiyuan@acoinfo.com>

package protocol

type MessageType byte

// Common VSOA conn type.
const (
	TypeServInfo    MessageType = iota //Shack hand between C/S
	TypeRPC                            //VSOA RPC call
	TypeSubscribe                      //VSOA subscribe
	TypeUnsubscribe                    //VSOA cannel subscribe
	TypePublish                        //VSOA Publish data to subscriber
	TypeDatagram                       //VSOA Datagram without resp
	TypeQosSetup                       //Setup Qos for VSOA
	TypePingEcho    MessageType = 0xff //VSOA internel ping call
)

// MethodText returns a text for the VSOA conn type code. It returns the empty
// string if the code is unknown.
func TypeText(code MessageType) string {
	switch code {
	case TypeServInfo:
		return "TYPE_SERVINFO"
	case TypeRPC:
		return "TYPE_RPC"
	case TypeSubscribe:
		return "TYPE_SUBSCRIBE"
	case TypeUnsubscribe:
		return "TYPE_UNSUBSCRIBE"
	case TypePublish:
		return "TYPE_PUBLISH"
	case TypeDatagram:
		return "TYPE_DATAGRAM"
	case TypeQosSetup:
		return "TYPE_QOS_SETUP"
	case TypePingEcho:
		return "TYPE_PING"
	default:
		return ""
	}
}
