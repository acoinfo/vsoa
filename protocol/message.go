// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: message.go Vehicle SOA protocol package.
//
// Author: Wang.yifan <wangyifan@acoinfo.com>
// Contributor: Cheng.siyuan <chengsiyuan@acoinfo.com>

package protocol

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/acoinfo/vsoa/utils"
)

// Constants for internal use in vsoa package
const (
	magic   byte = 0x9
	version byte = 0x2
	/* VSOA package length limits */
	HdrLength        = 10
	MaxMessageLength = 262144
	MaxDataLength    = (MaxMessageLength - HdrLength)
	/* VSOA quick channel length limit */
	MaxQMessageLength = 65507
)

var bufferPool = utils.NewLimitedPool(512, 4096)

func MagicNumber() byte {
	return magic | (version << 4)
}

var (
	ErrMessageTooLong = errors.New("message is too long")
	ErrMessageUnPad   = errors.New("raw message is not pad")
)

const (
	ServerError = "__vsoa_error__"
)

type QuickChannelFlag bool

const (
	ChannelQuick  QuickChannelFlag = true
	ChannelNormal QuickChannelFlag = false
)

type Message struct {
	*Header
	URL   []byte          // Server PATH for RPC
	Param json.RawMessage // JSON-encoded parameters
	Data  []byte          // Raw message data
}

func NewMessage() *Message {
	header := Header([HdrLength]byte{})
	header[0] = MagicNumber()
	return &Message{
		Header: &header,
	}
}

// Header represents the fixed-size header part of a Message
type Header [HdrLength]byte

func (h Header) MessageType() MessageType {
	return MessageType(h[1])
}

func (h Header) MessageTypeText() string {
	return TypeText(h.MessageType())
}

func (h *Header) SetMessageType(mt MessageType) {
	h[1] = byte(mt)
}

func (h Header) IsRPC() bool {
	return h[1] == 0x01
}

func (h Header) IsNoop() bool {
	return h[1] == 0xfe
}

func (h Header) IsPingEcho() bool {
	return h[1] == 0xff
}

func (h Header) IsServInfo() bool {
	return h[1] == 0x00
}

func (h Header) IsSubscribe() bool {
	return h[1] == 0x02
}

func (h Header) IsUnSubscribe() bool {
	return h[1] == 0x03
}

func (h Header) IsOneway() bool {
	return h[1] == 0x05 || h[1] == 0x04
}

func (h *Header) SetPingEcho() {
	h[1] = 0xff
}

func (h Header) IsReply() bool {
	return (h[2] & 0x01) != 0
}

func (h *Header) SetReply(r bool) {
	if r {
		h[2] |= 0x01
	} else {
		h[2] &^= 0x01
	}
}

func (h Header) IsValidTunid() bool {
	return (h[2]>>1)&0x01 != 0
}

func (h *Header) SetValidTunid() {
	h[2] |= 0x02
}

func (h Header) MessageRpcMethod() RpcMessageType {
	if h.IsRPC() {
		return RpcMessageType((h[2] >> 2) & 0x01)
	}
	return NoneRpc
}

func (h Header) MessageRpcMethodText() string {
	return RpcMethodText(h.MessageRpcMethod())
}

func (h *Header) SetMessageRpcMethod(t RpcMessageType) {
	if t == RpcMethodGet {
		h[2] &^= 0x04
	} else {
		h[2] |= 0x04
	}
}

func (h Header) StatusType() StatusType {
	return StatusType(h[3])
}

func (h Header) StatusTypeText() string {
	return StatusText(h.StatusType())
}

func (h *Header) SetStatusType(mt StatusType) {
	h[3] = byte(mt)
}

func (h Header) SeqNo() uint32 {
	return binary.BigEndian.Uint32(h[4:8])
}

func (h *Header) SetSeqNo(seq uint32) {
	binary.BigEndian.PutUint32(h[4:8], seq)
}

func (h Header) TunID() uint16 {
	return binary.BigEndian.Uint16(h[8:10])
}

func (h *Header) SetTunId(ti uint16) {
	binary.BigEndian.PutUint16(h[8:10], ti)
}

func (h Header) Check() bool { return h[0] == MagicNumber() }

// CloneHeader creates a new message with just the header copied
func (m Message) CloneHeader() *Message {
	header := *m.Header
	return &Message{Header: &header}
}

func (m *Message) Encode(quick QuickChannelFlag) ([]byte, error) {
	uLen, pLen, dLen := len(m.URL), len(m.Param), len(m.Data)
	total := HdrLength + 10 + uLen + pLen + dLen

	pad := (4 - (total & 3)) & 3
	total += pad

	max := MaxMessageLength
	if quick {
		max = MaxQMessageLength
	}

	if total > max {
		return nil, ErrMessageTooLong
	}

	m.Header[2] = (m.Header[2] & 0x3F) | byte(pad<<6)

	buf := make([]byte, total)
	copy(buf, m.Header[:])

	binary.BigEndian.PutUint16(buf[10:12], uint16(uLen))
	binary.BigEndian.PutUint32(buf[12:16], uint32(pLen))
	binary.BigEndian.PutUint32(buf[16:20], uint32(dLen))

	copy(buf[20:20+uLen], m.URL)
	copy(buf[20+uLen:20+uLen+pLen], m.Param)
	copy(buf[20+uLen+pLen:], m.Data)

	return buf, nil
}

// PutData returns a byte slice to the pool
func PutData(data *[]byte) {
	bufferPool.Put(data)
}

func (m *Message) Decode(r io.Reader) error {
	if _, err := io.ReadFull(r, m.Header[:]); err != nil || !m.Check() {
		return fmt.Errorf("invalid header")
	}

	var lens [10]byte
	if _, err := io.ReadFull(r, lens[:]); err != nil {
		return err
	}

	uLen := int(binary.BigEndian.Uint16(lens[0:2]))
	pLen := int(binary.BigEndian.Uint32(lens[2:6]))
	dLen := int(binary.BigEndian.Uint32(lens[6:10]))
	pad := int(m.Header[2] >> 6)

	total := HdrLength + 10 + uLen + pLen + dLen + pad
	if total&3 != 0 || total > MaxMessageLength {
		return ErrMessageTooLong
	}

	m.URL = utils.ResizeSliceSize(m.URL, uLen)
	if _, err := io.ReadFull(r, m.URL); err != nil {
		return err
	}

	m.Param = utils.ResizeSliceSize(m.Param, pLen)
	if _, err := io.ReadFull(r, m.Param); err != nil {
		return err
	}

	m.Data = utils.ResizeSliceSize(m.Data, dLen)
	if _, err := io.ReadFull(r, m.Data); err != nil {
		return err
	}

	if pad > 0 {
		io.CopyN(io.Discard, r, int64(pad))
	}

	return nil
}

// Reset clears the message while preserving allocated memory
func (m *Message) Reset() {
	m.Header[1] = 0
	m.Header[2] = 0
	m.Header[3] = 0
	binary.BigEndian.PutUint32(m.Header[4:8], 0)
	binary.BigEndian.PutUint16(m.Header[8:10], 0)
	m.URL = m.URL[:0]
	m.Param = m.Param[:0]
	m.Data = m.Data[:0]
}
