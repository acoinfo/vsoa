// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: rpcmethod.go Vehicle SOA protocal package.
//
// Author: Cheng.siyuan <chengsiyuan@acoinfo.com>

package protocol

type RpcMessageType byte

// Common VSOA RPC methods.
const (
	RpcMethodGet RpcMessageType = iota
	RpcMethodSet
	NoneRpc RpcMessageType = 0xee
)

// MethodText returns a text for the VSOA RPC method code. It returns the empty
// string if the code is unknown.
func RpcMethodText(code RpcMessageType) string {
	switch code {
	case RpcMethodGet:
		return "GET"
	case RpcMethodSet:
		return "SET"
	case NoneRpc:
		return "Means nothing"
	default:
		return ""
	}
}
