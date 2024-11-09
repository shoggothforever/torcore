package model

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func ParseFromFile(name string) ([]*BNode, error) {
	fd, err := os.OpenFile("./test.file", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	fmt.Println("open file successfully")
	reader := bufio.NewReader(fd)
	return RecursiveParse(reader)
}
func ParseFromString(str string) ([]*BNode, error) {
	reader := bufio.NewReader(strings.NewReader(str))
	return RecursiveParse(reader)
}

func RecursiveParse(reader *bufio.Reader) ([]*BNode, error) {
	nodes := make([]*BNode, 0)
	for {
		parse, err := BenParse(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Println("parse error", err)
			return nil, err
		} else if err == io.EOF {
			fmt.Println("EOF")
			break
		} else {
			nodes = append(nodes, parse)
		}
	}
	return nodes, nil
}
func BenParse(r io.Reader) (*BNode, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	b, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	node := new(BNode)
Parse:
	switch {
	case b[0] > '0' && b[0] < '9':
		{
			val := new(BStr)
			val.Decode(br, node)
		}
	case b[0] == 'i':
		{
			val := new(BInt)
			val.Decode(br, node)
		}
	case b[0] == 'l':
		{
			val := new(BList)
			val.Decode(br, node)
		}
	case b[0] == 'd':
		{
			dict := make(BDict)
			dict.Decode(br, node)

		}
	case b[0] == '\n':
		{
			br.ReadByte()
			b, err = br.Peek(1)
			if err != nil {
				return nil, err
			}
			goto Parse
		}
	}
	return node, nil
}
