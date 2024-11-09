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
	var p []*BObject
	err = Unmarshal(reader, &p)
	if err != nil {
		panic(err)
	}
}
func TestUnmarshalDict(t *testing.T) {
	//fd, err := os.OpenFile("./test.file", os.O_RDONLY, 0644)
	//if err != nil {
	//	panic(err)
	//}
	//defer fd.Close()
	//fmt.Println("open file successfully")
	//reader := bufio.NewReader(fd)
	//var p *TorrentFile
	//err = Unmarshal(reader, &p)
	//if err != nil {
	//	panic(err)
	//}
	var Bnode1 = []BObject{}
	var bp = make(BDict)
	type tt struct {
		Name  string `bencode:"name"`
		List  BList  `bencode:"list"`
		Listp *BList `bencode:"listp"`
	}
	ts := &tt{
		Name:  "test",
		List:  BList{},
		Listp: blistptr,
	}
	fmt.Println(reflect.ValueOf(reflect.New(reflect.TypeOf(Bnode1))))
	fmt.Println(reflect.ValueOf(reflect.New(reflect.TypeOf(ts).Elem())))
	Help(&bp)
	fmt.Println(reflect.ValueOf(ts).Elem().Field(0).CanSet())
	fv1 := reflect.ValueOf(ts).Elem().Field(1)
	fv2 := reflect.ValueOf(ts).Elem().Field(2)
	ft1 := fv1.Type()
	ft1 := fv1.Type()
	fmt.Println(ts)
}
func Help(b BObject) {
	fmt.Println(reflect.ValueOf(reflect.New(reflect.TypeOf(b).Elem())))

}
