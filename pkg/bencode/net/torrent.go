package net

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/shoggothforever/torcore/pkg/bencode/model"
	"net"
)

const (
	HashLength int    = 20
	Btag       string = "bencode"
	PeerPort   int    = 6666
	IpLen      int    = 4
	PortLen    int    = 2
	PeerLen    int    = IpLen + PortLen
)

type TorrentFile struct {
	Announce    string             `bencode:"announce"`
	Name        string             `bencode:"name"`
	PieceLength int                `bencode:"piecelength"`
	Length      int                `bencode:"length"`
	InfoHash    [HashLength]byte   `bencode:"infohash"`
	PieceHashes [][HashLength]byte `bencode:"pieces"`
}

type benInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	PieceLength int    `bencode:"piecelength"`
	Pieces      string `bencode:"pieces"`
}

type benTorrent struct {
	Announce string  `bencode:"announce"`
	Info     benInfo `bencode:"info"`
}
type PeerInfo struct {
	Ip   net.IP
	Port uint16
}

type TrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (i *benInfo) hash() ([HashLength]byte, error) {
	var buf bytes.Buffer
	model.MarshalBen(&buf, *i)
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *benInfo) splitPieceHashes() ([][HashLength]byte, error) {
	buf := []byte(i.Pieces)
	if len(buf)%HashLength != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / HashLength
	hashes := make([][HashLength]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*HashLength:(i+1)*HashLength])
	}
	return hashes, nil
}
func (bt *benTorrent) toTorrentFile() (*TorrentFile, error) {
	infoHash, err := bt.Info.hash()
	if err != nil {
		return nil, err
	}
	pieceHashes, err := bt.Info.splitPieceHashes()
	if err != nil {
		return nil, err
	}
	t := &TorrentFile{
		Announce:    bt.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bt.Info.PieceLength,
		Length:      bt.Info.Length,
		Name:        bt.Info.Name,
	}
	return t, nil
}
func buildPeerInfo(peers []byte) []PeerInfo {
	num := len(peers) / PeerLen
	if len(peers)%PeerLen != 0 {
		fmt.Println("Received malformed peers")
		return nil
	}
	infos := make([]PeerInfo, num)
	for i := 0; i < num; i++ {
		offset := i * PeerLen
		infos[i].Ip = net.IP(peers[offset : offset+IpLen])
		infos[i].Port = binary.BigEndian.Uint16(peers[offset+IpLen : offset+PeerLen])
	}
	return infos
}
