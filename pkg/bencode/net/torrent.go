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
	HashLength int    = 20
	Btag       string = "bencode"
	PeerPort   string = "13372"
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
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{PeerPort},
		"downloaded": []string{"0"},
		"uploaded":   []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

// 从种子文件获取peers信息，可能需要定时调用来更新peers信息
func (tf *TorrentFile) GetPeers(peerID [IDLEN]byte) ([]PeerInfo, error) {

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
	peers, err := tf.GetPeers(peerID)
	if err != nil {
		return err
	}

	torrent := Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()
	buf, err := torrent.Download(ctx)
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
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
