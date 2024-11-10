package net

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type PeerMsgType int

const (
	PChoked PeerMsgType = iota
	PUnChoked
	PInterested
	PNotInterest
	PHave
	PBitfield
	PRequest
	PPiece
	PCancel
)
const (
	PreMsgSizeBitLen = 1
	PreMsgSizeLen    = 19
	ReservedLen      = 8
	HandShakeMsgLen  = PreMsgSizeBitLen + PreMsgSizeLen + ReservedLen + HashLength + IDLEN
	DialTime         = 5 * time.Second
)

type PeerMsg struct {
	type_   PeerMsgType
	Payload []byte
}
type PeerInfo struct {
	Ip   net.IP
	Port uint16
}

type PeerConn struct {
	net.Conn
	Choked  bool
	Field   Bitfield
	peer    PeerInfo
	peerId  [IDLEN]byte
	infoSHA [HashLength]byte
}

func (c *PeerConn) ReadPeerMsg() (PeerMsg, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(c, buf)
	if err != nil {
		return PeerMsg{}, err
	}
	length := binary.BigEndian.Uint32(buf)
	buf = make([]byte, length)
	_, err = io.ReadFull(c, buf)
	if err != nil {
		return PeerMsg{}, err
	}
	return PeerMsg{
		type_:   PeerMsgType(buf[0]),
		Payload: buf[1:],
	}, nil
}
func (c *PeerConn) WritePeerMsg(m *PeerMsg) (int, error) {
	length := len(m.Payload) + 1
	lenBuf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(lenBuf[:4], uint32(length))
	lenBuf[4] = byte(m.type_)
	copy(lenBuf[5:], m.Payload)
	return c.Write(lenBuf)
}

// 对于have类型的peer消息，类型为PHave，而且payload固定为四字节
func GetHaveIndex(msg *PeerMsg) (int, error) {
	if msg.type_ != PHave || len(msg.Payload) != 4 {
		return -1, errors.New("wrong form of HAVE peer message ")
	}
	index := binary.BigEndian.Uint32(msg.Payload)
	return int(index), nil
}
func NewConn(peer PeerInfo, infoSha [HashLength]byte, peerID [IDLEN]byte) (*PeerConn, error) {
	addr := net.JoinHostPort(peer.Ip.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", addr, DialTime)
	if err != nil {
		return nil, err
	}
	err = handShake(conn, infoSha, peerID)
	if err != nil {
		fmt.Println("handshake failed")
		conn.Close()
		return nil, err
	}
	pconn := &PeerConn{
		Conn:    conn,
		Choked:  true,
		peer:    peer,
		peerId:  peerID,
		infoSHA: infoSha,
	}
	return pconn, nil
}
func NewPeerMsg(index, offset, length int) PeerMsg {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], uint32(index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(offset))
	binary.BigEndian.PutUint32(buf[8:12], uint32(length))
	return PeerMsg{PRequest, buf}
}
func handShake(conn net.Conn, infoSha [HashLength]byte, peerID [IDLEN]byte) error {
	err := conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return err
	}
	defer conn.SetDeadline(time.Time{})
	req := newHandShakeReq(infoSha, peerID)
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

type handShakeMsg struct {
	PreMsg  string
	infoSha [HashLength]byte
	peerID  [IDLEN]byte
}

func newHandShakeReq(infoSha [HashLength]byte, peerID [IDLEN]byte) handShakeMsg {
	return handShakeMsg{
		PreMsg:  "BitTorrent protocol",
		infoSha: infoSha,
		peerID:  peerID,
	}
}
func writeHandShake(w io.Writer, req handShakeMsg) (int, error) {
	buf := make([]byte, PreMsgSizeBitLen+len(req.PreMsg)+ReservedLen+HashLength+IDLEN)
	buf[0] = 0x13
	n := 1
	n += copy(buf[n:n+len(req.PreMsg)], req.PreMsg)
	n += copy(buf[n:n+ReservedLen], make([]byte, ReservedLen))
	n += copy(buf[n:n+HashLength], req.infoSha[:])
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
	buf = make([]byte, PreMsgSizeLen+ReservedLen+HashLength+IDLEN)
	_, err = io.ReadFull(w, buf)
	if err != nil {
		return handShakeMsg{}, err
	}
	st := 0
	msg := handShakeMsg{}
	msg.PreMsg = string(buf[st : st+PreMsgSizeLen])
	st += PreMsgSizeLen + ReservedLen
	msg.infoSha = [HashLength]byte(buf[st : st+HashLength])
	st += HashLength
	msg.peerID = [IDLEN]byte(buf[st : st+IDLEN])
	return msg, nil
}
func copyPieceData(index int, buf []byte, msg *PeerMsg) (int, error) {
	if msg.type_ != PPiece {
		return -1, errors.New("Piece message type illegal")
	}
	if len(msg.Payload) < 8 {
		return -1, errors.New("Piece's payload length illegal")
	}
	id := binary.BigEndian.Uint32(buf[:4])
	if int(id) != index {
		return -1, errors.New("Piece's id illegal")
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
