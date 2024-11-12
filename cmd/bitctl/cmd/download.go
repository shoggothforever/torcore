/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	mt "github.com/shoggothforever/torcore/pkg/bencode/net"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var fileName string
var outputPath string
var deadline int
var pre bool

// NewMarshalCmd represents the marshal command
func NewDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "input the torrent file,",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Aliases: []string{"dl"},
		Run:     DownloadFunc,
	}
	cmd.Flags().StringVarP(&fileName, "file", "f", "filename", "input fileName that contain bencode Object text to marshal into bencode text")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "./output", "the path where files downloaded into")
	cmd.Flags().IntVarP(&deadline, "deadline", "d", -1, "limit max download time ")
	cmd.Flags().BoolVarP(&pre, "prelude", "p", false, "get a glimpse of torrent")
	return cmd
}
func init() {
	rootCmd.AddCommand(NewDownloadCmd())
}
func DownloadFunc(cmd *cobra.Command, args []string) {
	fd, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	fmt.Println("open file:", fileName, " successfully")
	r := bufio.NewReader(fd)
	t, err := mt.UnmarshalTorrentFile(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("get torrent file, length: ", t.FileLen)
	fmt.Println("get pre bool ", pre)
	fmt.Println(t.InfoSHA)
	if !pre {
		dur := (time.Duration)(deadline) * time.Second
		err = t.DownloadToFile(outputPath, dur)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
