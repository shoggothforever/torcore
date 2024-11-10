package model

import (
	"bytes"
	"crypto/sha1"
	"io"
)

func ParseFile(r io.Reader) (*TorrentFile, error) {
	bt := new(benTorrent)
	err := Unmarshal(r, bt)
	if err != nil {
		return nil, err
	}
	ret := new(TorrentFile)
	ret.Name = bt.Info.Name
	ret.Length = bt.Info.Length
	ret.Announce = bt.Announce
	ret.PieceLength = bt.Info.PieceLength

	buf := new(bytes.Buffer)
	wlen := Marshal(buf, bt.Info)
	if wlen == 0 {
		return nil, ErrParse
	}
	ret.InfoHash = sha1.Sum(buf.Bytes())

	btp := []byte(bt.Info.Pieces)
	cnt := len(btp) / bt.Info.PieceLength
	hashes := make([][HashLength]byte, cnt)
	for i := 0; i < cnt; i++ {
		copy(hashes[i][:], btp[i*bt.Info.PieceLength:(i+1)*bt.Info.PieceLength])
	}
	ret.PieceHashes = hashes
	return ret, nil
}
