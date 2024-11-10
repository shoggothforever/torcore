package model

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

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
	err = UnmarshalBen(reader, p)
	if err != nil {
		panic(err)
	}
	fmt.Println("unmarshal dict successfully")
	fmt.Println(p)
}
func TestMarshalList(t *testing.T) {
	fd, err := os.OpenFile("./mTest", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	fmt.Println("open file successfully")
	//Writer := bufio.NewReader(fd)
	//n := MarshalBen(Writer)

}
func Help(b BObject) {
	fmt.Println(reflect.ValueOf(reflect.New(reflect.TypeOf(b).Elem())))
}

type User struct {
	Name string `bencode:"name"`
	Age  int    `bencode:"age"`
}

type Role struct {
	Id   int
	User `bencode:"user"`
}

type Score struct {
	User  `bencode:"user"`
	Value []int `bencode:"value"`
}

type Team struct {
	Name   string `bencode:"name"`
	Size   int    `bencode:"size"`
	Member []User `bencode:"member"`
}

func TestMarshalBasic(t *testing.T) {
	buf := new(bytes.Buffer)
	str := "abc"
	len := MarshalBen(buf, str)
	assert.Equal(t, 5, len)
	assert.Equal(t, "3:abc", buf.String())

	buf.Reset()
	val := 199
	len = MarshalBen(buf, val)
	assert.Equal(t, 5, len)
	assert.Equal(t, "i199e", buf.String())
}

func TestUnmarshalList1(t *testing.T) {
	str := "li85ei90ei95ee"
	l := &[]int{}
	UnmarshalBen(bytes.NewBufferString(str), l)
	assert.Equal(t, []int{85, 90, 95}, *l)

	buf := new(bytes.Buffer)
	length := MarshalBen(buf, l)
	assert.Equal(t, len(str), length)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalUser(t *testing.T) {
	str := "d4:name6:archer3:agei29ee"
	u := &User{}
	UnmarshalBen(bytes.NewBufferString(str), u)
	assert.Equal(t, "archer", u.Name)
	assert.Equal(t, 29, u.Age)

	buf := new(bytes.Buffer)
	length := MarshalBen(buf, u)
	fmt.Println(*u)
	assert.Equal(t, len(str), length)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalRole(t *testing.T) {
	str := "d2:idi1e4:userd4:name6:archer3:agei29eee"
	r := &Role{}
	UnmarshalBen(bytes.NewBufferString(str), r)
	assert.Equal(t, 1, r.Id)
	assert.Equal(t, "archer", r.Name)
	assert.Equal(t, 29, r.Age)

	buf := new(bytes.Buffer)
	length := MarshalBen(buf, r)
	assert.Equal(t, len(str), length)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalScore(t *testing.T) {
	str := "d4:userd4:name6:archer3:agei29ee5:valueli80ei85ei90eee"
	s := &Score{}
	UnmarshalBen(bytes.NewBufferString(str), s)
	assert.Equal(t, "archer", s.Name)
	assert.Equal(t, 29, s.Age)
	assert.Equal(t, []int{80, 85, 90}, s.Value)

	buf := new(bytes.Buffer)
	length := MarshalBen(buf, s)
	assert.Equal(t, len(str), length)
	assert.Equal(t, str, buf.String())
}

func TestUnmarshalTeam(t *testing.T) {
	str := "d4:name3:ace4:sizei2e6:memberld4:name6:archer3:agei29eed4:name5:nancy3:agei31eeee"
	//str := "d4:name6:archer4:sizei10e6:memberld4:name2:aa3:agei20eeee"
	team := &Team{}
	UnmarshalBen(bytes.NewBufferString(str), team)
	fmt.Println(team)
	assert.Equal(t, "ace", team.Name)
	assert.Equal(t, 2, team.Size)

	buf := new(bytes.Buffer)
	length := MarshalBen(buf, team)
	fmt.Println(buf)
	assert.Equal(t, len(str), length)
	assert.Equal(t, str, buf.String())
}
