package net

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type MsgID int

const (
	// MsgChoke chokes the receiver
	MsgChoke MsgID = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke MsgID = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested MsgID = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested MsgID = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave MsgID = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield MsgID = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest MsgID = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece MsgID = 7
	// MsgCancel cancels a request
	MsgCancel MsgID = 8
)

func (m *Message) name() string {
	if m == nil {
		return "KeepAlive"
	}
	switch m.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}
}
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 0)
	}
	length := len(m.Payload) + 1
	buf := make([]byte, length+4)
	binary.BigEndian.PutUint32(buf[:4], uint32(length))
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}
func (m *Message) String() string {
	if m == nil {
		return m.name()
	}
	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}

const (
	PreMsgSizeBitLen = 1
	PreMsgSizeLen    = 19
	ReservedLen      = 8
	HandShakeMsgLen  = PreMsgSizeBitLen + PreMsgSizeLen + ReservedLen + SHALEN + IDLEN
	DialTime         = 5 * time.Second
)

// p2p消息
type Message struct {
	ID      MsgID
	Payload []byte
}

// 对于have类型的peer消息，类型为PHave，而且payload固定为四字节
func GetHaveIndex(msg *Message) (int, error) {
	if msg.ID != MsgHave || len(msg.Payload) != 4 {
		return -1, errors.New("wrong form of HAVE peer message ")
	}
	index := binary.BigEndian.Uint32(msg.Payload)
	return int(index), nil
}

// 创建各种类型的message
// MsgHave
func NewHaveMessage(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

// basicMessage
func NewMessage(ID MsgID) *Message {
	return &Message{ID: ID}
}

// Interested
func NewInterestedMessage() *Message {
	return &Message{ID: MsgInterested}
}

// Unchoke
func NewUnchokeMessage() *Message {
	return &Message{ID: MsgUnchoke}
}

// MsgRequest
func NewRequestMessage(index, offset, length int) Message {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], uint32(index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(offset))
	binary.BigEndian.PutUint32(buf[8:12], uint32(length))
	return Message{MsgRequest, buf}
}

// 对等实体之间握手建立连接
func handShake(conn net.Conn, infoSha [SHALEN]byte, peerID [IDLEN]byte) error {
	err := conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return err
	}
	defer conn.SetDeadline(time.Time{})
	req := newHandShakeMsg(infoSha, peerID)
	_, err = writeHandShake(conn, req)
	if err != nil {
		return err
	}
	res, err := readHandShake(conn)
	if err != nil {
		return err
	}
	if !bytes.Equal(req.infoSha[:], res.infoSha[:]) {
		return errors.New("invalid info_sha")
	}
	return nil
}

// 握手报文
type handShakeMsg struct {
	PreMsg  string
	infoSha [SHALEN]byte
	peerID  [IDLEN]byte
}

// 新建握手报文信息
func newHandShakeMsg(infoSha [SHALEN]byte, peerID [IDLEN]byte) handShakeMsg {
	return handShakeMsg{
		PreMsg:  "BitTorrent protocol",
		infoSha: infoSha,
		peerID:  peerID,
	}
}
func writeHandShake(w io.Writer, req handShakeMsg) (int, error) {
	buf := make([]byte, PreMsgSizeBitLen+len(req.PreMsg)+ReservedLen+SHALEN+IDLEN)
	buf[0] = 0x13
	n := 1
	n += copy(buf[n:n+len(req.PreMsg)], req.PreMsg)
	n += copy(buf[n:n+ReservedLen], make([]byte, ReservedLen))
	n += copy(buf[n:n+SHALEN], req.infoSha[:])
	n += copy(buf[n:n+IDLEN], req.peerID[:])
	return w.Write(buf)
}
func readHandShake(w io.Reader) (handShakeMsg, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(w, buf)
	if err != nil {
		return handShakeMsg{}, err
	}
	if buf[0] != PreMsgSizeLen {
		return handShakeMsg{}, errors.New("invalid packet")
	}
	buf = make([]byte, PreMsgSizeLen+ReservedLen+SHALEN+IDLEN)
	_, err = io.ReadFull(w, buf)
	if err != nil {
		return handShakeMsg{}, err
	}
	st := 0
	msg := handShakeMsg{}
	msg.PreMsg = string(buf[st : st+PreMsgSizeLen])
	st += PreMsgSizeLen + ReservedLen
	msg.infoSha = [SHALEN]byte(buf[st : st+SHALEN])
	st += SHALEN
	msg.peerID = [IDLEN]byte(buf[st : st+IDLEN])
	return msg, nil
}
func copyPieceData(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return -1, errors.New("Piece message type illegal")
	}
	if len(msg.Payload) < 8 {
		return -1, errors.New("Piece's payload length illegal")
	}
	id := binary.BigEndian.Uint32(msg.Payload[:4])
	if int(id) != index {
		return -1, fmt.Errorf("%d!=%d %s", int(id), index, errors.New("Piece's id illegal").Error())
	}
	offset := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if offset >= len(buf) {
		return 0, fmt.Errorf("offset too high. %d >= %d", offset, len(buf))
	}
	data := msg.Payload[8:]
	if offset+len(data) > len(buf) {

		return 0, fmt.Errorf("data too large [%d] for offset %d with length %d", len(data), offset, len(buf))
	}
	copy(buf[offset:], data)
	return len(data), nil
}
