package net

import (
	"fmt"
	"github.com/shoggothforever/torcore/pkg/bencode/model"
	"github.com/shoggothforever/torcore/pkg/bencode/util"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func ParseTorrentFile(r io.Reader) (*TorrentFile, error) {
	bt := new(benTorrent)
	err := model.UnmarshalBen(r, bt)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("\n%+v\n", *bt)
	return bt.toTorrentFile()

}
func (tf *TorrentFile) BuildTrackerUrl() (string, error) {
	peerID := util.GeneratePeerID("dsm")
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{"13372"},
		"downloaded": []string{"0"},
		"uploaded":   []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}
func (tf *TorrentFile) GetPeers() ([]PeerInfo, error) {
	url, err := tf.BuildTrackerUrl()
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
