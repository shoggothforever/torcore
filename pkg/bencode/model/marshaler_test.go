package model

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"testing"
)

type BStruct struct {
	type_ string
	value string
}

func TestUnmarshalList(t *testing.T) {
	fd, err := os.OpenFile("./test.file", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	fmt.Println("open file successfully")
	reader := bufio.NewReader(fd)
	var p BList
	err = Unmarshal(reader, &p)
	if err != nil {
		panic(err)
	}
	fmt.Println("unmarshal successfully")
	PrintBobj(&p, "")
}
func TestUnmarshalDict(t *testing.T) {
	fd, err := os.OpenFile("./test.file", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	fmt.Println("open file successfully")
	reader := bufio.NewReader(fd)

	type testMp struct {
		Key1 int    `bencode:"key1"`
		Key2 string `bencode:"key2"`
	}
	p := &testMp{
		Key1: 123,
		Key2: "no",
	}
	err = Unmarshal(reader, p)
	if err != nil {
		panic(err)
	}
	fmt.Println("unmarshal dict successfully")
	fmt.Println(p)
}
func Help(b BObject) {
	fmt.Println(reflect.ValueOf(reflect.New(reflect.TypeOf(b).Elem())))
}
