package main

import (
	"github.com/shoggothforever/torcore/pkg/bencode/net"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]
	maxTime := os.Args[3]
	tf, err := net.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	mt, err := strconv.Atoi(maxTime)
	if err != nil {
		log.Fatal(err)
	}
	err = tf.DownloadToFile(outPath, time.Duration(mt)*time.Second)
	if err != nil {
		log.Fatal(err)
	}
}
