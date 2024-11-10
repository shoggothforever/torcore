package net

import (
	"fmt"
	"os"
	"testing"
)

func TestParseTorrentFile(t *testing.T) {
	r, _ := os.OpenFile("pl.torrent", os.O_RDONLY, 0644)
	file, err := UnmarshalTorrentFile(r)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print(file)
}

func TestTorrentFile_GetPeers(t *testing.T) {
	r, _ := os.OpenFile("pl.torrent", os.O_RDONLY, 0644)
	file, err := UnmarshalTorrentFile(r)
	if err != nil {
		t.Fatal(err)
		return
	}
	peers, err := file.GetPeers()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(peers)
}
func TestTorrentDBFile_GetPeers(t *testing.T) {
	r, _ := os.OpenFile("debian-iso.torrent", os.O_RDONLY, 0644)
	file, err := UnmarshalTorrentFile(r)
	if err != nil {
		t.Fatal(err)
		return
	}
	peers, err := file.GetPeers()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(peers)
}
