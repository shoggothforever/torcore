package net

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

func buildPeerInfo(peers []byte) []PeerInfo {
	num := len(peers) / PeerLen
	if len(peers)%PeerLen != 0 {
		fmt.Println("Received malformed peers")
		return nil
	}
	infos := make([]PeerInfo, num)
	for i := 0; i < num; i++ {
		offset := i * PeerLen
		infos[i].Ip = peers[offset : offset+IpLen]
		infos[i].Port = binary.BigEndian.Uint16(peers[offset+IpLen : offset+PeerLen])
	}
	return infos
}

type Bitfield []byte

func (field Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(field) {
		return false
	}
	return field[byteIndex]>>uint(7-offset)&1 != 0
}

func (field Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(field) {
		return
	}
	field[byteIndex] |= 1 << uint(7-offset)
}

func (field Bitfield) String() string {
	str := "piece# "
	for i := 0; i < len(field)*8; i++ {
		if field.HasPiece(i) {
			str = str + strconv.Itoa(i) + " "
		}
	}
	return str
}

// p2p对等实体网络信息
type PeerInfo struct {
	Ip   net.IP
	Port uint16
}

// p2p对等实体连接信息
type PeerConn struct {
	net.Conn
	Choked  bool
	Field   Bitfield
	peer    PeerInfo
	peerId  [IDLEN]byte
	infoSHA [HashLength]byte
}

// 创建一个对等实体的连接
// infoSha 用于校验，peerID在一次下载中唯一
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
	c := &PeerConn{
		Conn:    conn,
		Choked:  true,
		peer:    peer,
		peerId:  peerID,
		infoSHA: infoSha,
	}
	err = c.ReadBitFieldMessage()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// 从网络连接中获取p2p信息
func (c *PeerConn) ReadMessage() (Message, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(c, buf)
	if err != nil {
		return Message{}, err
	}
	length := binary.BigEndian.Uint32(buf)
	buf = make([]byte, length)
	_, err = io.ReadFull(c, buf)
	if err != nil {
		return Message{}, err
	}
	return Message{
		ID:      MsgID(buf[0]),
		Payload: buf[1:],
	}, nil
}

// 向网络连接写入p2p信息
func (c *PeerConn) WriteMessage(m *Message) (int, error) {
	return c.Write(m.Serialize())
}
func (c *PeerConn) ReadBitFieldMessage() error {
	c.SetDeadline(time.Now().Add(5 * time.Second))
	defer c.SetDeadline(time.Time{})

	msg, err := c.ReadMessage()
	if err != nil {
		return err
	}
	if msg.ID != MsgBitfield {
		return fmt.Errorf("expected bitfield, get " + strconv.Itoa(int(msg.ID)))
	}
	fmt.Println("fill bitfield : " + c.peer.Ip.String())
	c.Field = msg.Payload
	return nil
}
