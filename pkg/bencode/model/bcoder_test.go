package model

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

var bintptr *BInt = new(BInt)
var bstrptr *BStr = new(BStr)
var blistptr *BList = new(BList)
var bdictptr *BDict = new(BDict)

func TestEncodeBNode(t *testing.T) {
	*bintptr = -233
	*bstrptr = "str"
	*blistptr = BList{bintptr, bstrptr, bdictptr, bintptr}
	*bdictptr = map[string]BObject{"key1": bintptr, "key2": bstrptr}
	var Bnode1 = BNode{
		type_: BINT,
		data:  bintptr,
	}
	fd, err := os.OpenFile("./test.file", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	writer := bufio.NewWriter(fd)
	n, err := Encode(&Bnode1, writer)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("encode length ", n)
	var Bnode2 = BNode{
		type_: BSTR,
		data:  bstrptr,
	}
	n, err = Encode(&Bnode2, writer)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("encode length ", n)
	var Bnode3 = BNode{
		type_: BLIST,
		data:  blistptr,
	}
	n, err = Encode(&Bnode3, writer)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("encode length ", n)
	var Bnode4 = BNode{
		BDICT,
		bdictptr,
	}
	n, err = Encode(&Bnode4, writer)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("encode length ", n)
}
func TestDecodeBNode(t *testing.T) {
	nodes, err := DecodeFromFile("./test.file")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(nodes))
	for _, v := range nodes {
		fmt.Println(*v)
	}
}
