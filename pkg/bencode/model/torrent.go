package model

import (
	"fmt"
)

const HashLength = 20

type TorrentFile struct {
	Announce    string             `bencode:"announce"`
	InfoHash    [HashLength]byte   `bencode:"infohash"`
	PieceHashes [][HashLength]byte `bencode:"pieces"`
	PieceLength int                `bencode:"piecelength"`
	Length      int                `bencode:"length"`
	Name        string             `bencode:"name"`
}

type bencodeInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	PieceLength int    `bencode:"piecelength"`
	Pieces      string `bencode:"pieces"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

//func (i *bencodeInfo) hash() ([HashLength]byte, error) {
//	var buf bytes.Buffer
//	err := bencode.Marshal(&buf, *i)
//	if err != nil {
//		return [HashLength]byte{}, err
//	}
//	h := sha1.Sum(buf.Bytes())
//	return h, nil
//}

func (i *bencodeInfo) splitPieceHashes() ([][HashLength]byte, error) {
	hashLen := HashLength // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][HashLength]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

//func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
//	infoHash, err := bto.Info.hash()
//	if err != nil {
//		return TorrentFile{}, err
//	}
//	pieceHashes, err := bto.Info.splitPieceHashes()
//	if err != nil {
//		return TorrentFile{}, err
//	}
//	t := TorrentFile{
//		Announce:    bto.Announce,
//		InfoHash:    infoHash,
//		PieceHashes: pieceHashes,
//		PieceLength: bto.Info.PieceLength,
//		Length:      bto.Info.Length,
//		Name:        bto.Info.Name,
//	}
//	return t, nil
//}
