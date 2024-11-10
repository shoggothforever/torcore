package net

import (
	"fmt"
	"os"
	"testing"
)

func TestParseTorrentFile(t *testing.T) {
	r, _ := os.OpenFile("xxx.torrent", os.O_RDONLY, 0644)
	file, err := ParseTorrentFile(r)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print(file)
}
