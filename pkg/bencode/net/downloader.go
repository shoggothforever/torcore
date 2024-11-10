package net

import (
	"bytes"
	"crypto/sha1"
	"fmt"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}
type pieceProgress struct {
	index      int
	client     *PeerConn
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.ReadMessage() // this call blocks
	if err != nil {
		return err
	}
	switch msg.ID {
	case MsgUnchoke:
		state.client.Choked = false
	case MsgChoke:
		state.client.Choked = true
	case MsgHave:
		index, err := GetHaveIndex(&msg)
		if err != nil {
			return err
		}
		state.client.Field.SetPiece(index)
	case MsgPiece:
		n, err := copyPieceData(state.index, state.buf, &msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}
func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}
