package net

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/shoggothforever/torcore/pkg/bencode/model"
)

const (
	HashLength int    = 20
	Btag       string = "bencode"
	PeerPort   int    = 6666
	IpLen      int    = 4
	PortLen    int    = 2
	PeerLen    int    = IpLen + PortLen
	IDLEN      int    = 20
)

type TorrentFile struct {
	Announce    string             `bencode:"announce"`
	Name        string             `bencode:"name"`
	PieceLength int                `bencode:"piece length"`
	Length      int                `bencode:"length"`
	InfoHash    [HashLength]byte   `bencode:"info hash"`
	PieceHashes [][HashLength]byte `bencode:"pieces"`
}

type benInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Private     string `bencode:"private"`
	Url         string `bencode:"url"`
}

type benTorrent struct {
	Announce string  `bencode:"announce"`
	Info     benInfo `bencode:"info"`
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

func (bi *benInfo) splitPieceHashes() ([][HashLength]byte, error) {
	buf := []byte(bi.Pieces)
	numHashes := len(buf) / HashLength
	hashes := make([][HashLength]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*HashLength:min((i+1)*HashLength, len(buf))])
	}
	return hashes, nil
}
func (bt *benTorrent) toTorrentFile() (*TorrentFile, error) {
	infoHash, err := bt.Info.hash()
	if err != nil {
		return nil, err
	}
	fmt.Println("calc info hash ", infoHash)
	pieceHashes, err := bt.Info.splitPieceHashes()
	if err != nil {
		return nil, err
	}
	//fmt.Println("calc piece hashed  ", pieceHashes)
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
