package net

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/shoggothforever/torcore/pkg/bencode/model"
	"github.com/shoggothforever/torcore/pkg/bencode/util"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	SHALEN   int    = 20
	Btag     string = "bencode"
	PeerPort string = "6666"
	IpLen    int    = 4
	PortLen  int    = 2
	PeerLen  int    = IpLen + PortLen
	IDLEN    int    = 20
)

type TorrentFile struct {
	Announce string
	InfoSHA  [SHALEN]byte
	FileName string
	FileLen  int
	PieceLen int
	PieceSHA [][SHALEN]byte
}

type benInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type benTorrent struct {
	Announce string  `bencode:"announce"`
	Info     benInfo `bencode:"info"`
}

type TrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (bi *benInfo) hash() ([SHALEN]byte, error) {
	var buf bytes.Buffer
	model.MarshalBen(&buf, *bi)
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (bi *benInfo) splitPieceSHA() ([][SHALEN]byte, error) {
	buf := []byte(bi.Pieces)
	numHashes := len(buf) / SHALEN
	hashes := make([][SHALEN]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*SHALEN:min((i+1)*SHALEN, len(buf))])
	}
	return hashes, nil
}
func (bt *benTorrent) toTorrentFile() (*TorrentFile, error) {
	infoHash, err := bt.Info.hash()
	if err != nil {
		return nil, err
	}
	fmt.Println("calc info hash ", infoHash)
	PieceSHA, err := bt.Info.splitPieceSHA()
	if err != nil {
		return nil, err
	}
	//fmt.Println("calc piece hashed  ", PieceSHA)
	t := &TorrentFile{
		Announce: bt.Announce,
		InfoSHA:  infoHash,
		PieceSHA: PieceSHA,
		PieceLen: bt.Info.PieceLength,
		FileLen:  bt.Info.Length,
		FileName: bt.Info.Name,
	}
	return t, nil
}
func UnmarshalTorrentFile(r io.Reader) (*TorrentFile, error) {
	bt := new(benTorrent)
	err := model.UnmarshalBen(r, bt)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("\n%+v\n", *bt)
	return bt.toTorrentFile()

}

// 获取资源追踪站点网址信息
func (tf *TorrentFile) buildTrackerUrl(peerID [IDLEN]byte) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(tf.InfoSHA[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{PeerPort},
		"downloaded": []string{"0"},
		"uploaded":   []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.FileLen)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

// 从种子文件获取peers信息，可能需要定时调用来更新peers信息
func (tf *TorrentFile) getPeers(peerID [IDLEN]byte) ([]PeerInfo, error) {
	url, err := tf.buildTrackerUrl(peerID)
	if err != nil {
		fmt.Println("failed to build tracker url: ", err.Error())
		return nil, err
	}
	fmt.Println("get tracker url " + url)
	cli := &http.Client{Timeout: 15 * time.Second}
	resp, err := cli.Get(url)
	if err != nil {
		fmt.Println("failed to connect to Tracker: ", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	trsp := new(TrackerResp)
	err = model.UnmarshalBen(resp.Body, trsp)
	if err != nil {
		return nil, err
	}
	return buildPeerInfo([]byte(trsp.Peers)), nil
}

// DownloadToFile downloads a torrent and writes it to a file
func (tf *TorrentFile) DownloadToFile(path string, maxTime time.Duration) error {
	peerID := util.GeneratePeerID("dsm")
	peers, err := tf.getPeers(peerID)
	if err != nil {
		return err
	}
	torrent := Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoSHA:     tf.InfoSHA,
		PieceSHA:    tf.PieceSHA,
		PieceLength: tf.PieceLen,
		Length:      tf.FileLen,
		Name:        tf.FileName,
	}
	ctx := context.Background()
	var cancel context.CancelFunc
	if maxTime > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), maxTime)
		defer cancel()
	}
	buf, err := torrent.download(ctx)
	if err != nil {
		return err
	}

	outFile, err := os.Create(path + tf.FileName)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// Open parses a torrent file
func Open(path string) (*TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bt := new(benTorrent)
	err = model.UnmarshalBen(file, bt)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("\n%+v\n", *bt)
	return bt.toTorrentFile()
}
