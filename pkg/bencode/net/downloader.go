package net

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

const ReceiveGNums = 16

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

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []*PeerInfo
	PeerID      [IDLEN]byte
	InfoSHA     [SHALEN]byte
	PieceSHA    [][SHALEN]byte
	PieceLength int
	Length      int
	Name        string
	m           sync.Mutex
	mp          map[string]struct{}
	wg          sync.WaitGroup
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
		state.client.BitField.SetPiece(index)
	case MsgPiece:
		n, err := copyPieceData(state.index, state.buf, &msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		//fmt.Println("get peers pieces ", n)
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

// 与对等实体建立连接，从workQueue中获取需要的工作，然后开始真正的下载工作，最后将结果写入结果队列
func (t *Torrent) startDownload(ctx context.Context, peer *PeerInfo, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := NewConn(peer, t.InfoSHA, t.PeerID)
	if err != nil {
		return
	}
	fmt.Println("connect successfully")
	defer c.Conn.Close()
	c.SendBasicMessage(MsgInterested)
	c.SendBasicMessage(MsgUnchoke)

	for pw := range workQueue {
		if !c.BitField.HasPiece(pw.index) {
			workQueue <- pw
			continue
		}
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}
		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}
		//fmt.Println("check right")
		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}
func attemptDownloadPiece(c *PeerConn, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}
	err := c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		return nil, err
	}
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}
				err = c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err = state.readMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

func (t *Torrent) download(ctx context.Context, tf *TorrentFile) ([]byte, error) {
	workerQueue := make(chan *pieceWork, len(t.PieceSHA))
	ResQueue := make(chan *pieceResult, len(t.PieceSHA))
	for index, hash := range t.PieceSHA {
		length := t.calculatePieceSize(index)
		workerQueue <- &pieceWork{index, hash, length}
	}

	go t.updatePeersConn(ctx, tf, workerQueue, ResQueue)

	buf := make([]byte, t.Length)
	donePieces := atomic.Int64{}
	donePieces.Store(0)
	for i := 0; i < ReceiveGNums; i++ {
		t.wg.Add(1)
		go func() {
			tk := time.NewTicker(time.Second)
			defer t.wg.Done()
			for int(donePieces.Load()) < len(t.PieceSHA) {
				select {
				case res := <-ResQueue:
					begin, end := t.calculateBoundsForPiece(res.index)
					copy(buf[begin:end], res.buf)
					donePieces.Add(1)
					percent := float64(donePieces.Load()) / float64(len(t.PieceSHA)) * 100
					numWorkers := runtime.NumGoroutine() - 1 - ReceiveGNums // subtract 1 for main thread
					log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
				case <-tk.C:
					continue
				case <-ctx.Done():
					break
				}
			}
		}()
	}
	t.wg.Wait()
	close(ResQueue)
	close(workerQueue)
	return buf, nil
}
func (t *Torrent) updatePeersConn(ctx context.Context, tf *TorrentFile, workerQueue chan *pieceWork, ResQueue chan *pieceResult) {
	tk := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-tk.C:
			peers, err := tf.getPeers(t.PeerID)
			if err != nil {
				return
			}
			for _, peer := range peers {
				pi := net.JoinHostPort(peer.Ip.String(), strconv.Itoa(int(peer.Port)))
				if _, ok := t.mp[pi]; !ok {
					t.mp[pi] = struct{}{}
					t.Peers = append(t.Peers, peer)
					go t.startDownload(ctx, peer, workerQueue, ResQueue)
				}
			}
		}
	}

}
func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}
