// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: status.go Vehicle SOA protocal package.
//
// Author: Cheng.siyuan <chengsiyuan@acoinfo.com>

package protocol

type StatusType byte

// VSOA RPC status codes as registered with ACOINFO.
const (
	StatusSuccess       StatusType = iota // VSOA 1.0
	StatusPassword                        // VSOA 1.0
	StatusArguments                       // VSOA 1.0
	StatusInvalidUrl                      // VSOA 1.0
	StatusNoResponding                    // VSOA 1.0
	StatusNoPermissions                   // VSOA 1.0
	StatusNoMemory                        // VSOA 1.0
)

// StatusText returns a text for the VSOA RPC status code. It returns the empty
// string if the code is unknown.
func StatusText(code StatusType) string {
	switch code {
	case StatusSuccess:
		return "Success"
	case StatusPassword:
		return "Password"
	case StatusInvalidUrl:
		return "Invalid URL"
	case StatusNoResponding:
		return "No responding"
	case StatusNoPermissions:
		return "No permissions"
	case StatusNoMemory:
		return "No Memory"
	default:
		return ""
	}
}
