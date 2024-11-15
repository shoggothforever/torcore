package model

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Btype uint8

const (
	BINVALID Btype = iota
	BSTR
	BINT
	BLIST
	BDICT
)
const BTag = "bencode"

type BObject interface {
	Type() Btype
	Encode(writer io.Writer) (int, error)
	Decode(br *bufio.Reader, node *BNode)
}

var bint BObject = (*BInt)(nil)
var bstr BObject = (*BStr)(nil)
var bList BObject = (*BList)(nil)
var bDict BObject = (*BDict)(nil)

type BInt int

func (o *BInt) Type() Btype {
	return BINT
}

func (o *BInt) Encode(writer io.Writer) (int, error) {
	bw, ok := writer.(*bufio.Writer)
	if !ok {
		return -1, ErrTyp
	}
	wlen := 2
	num := int(*o)
	bw.WriteString("i")
	n, err := bw.WriteString(strconv.Itoa(num))
	if err != nil {
		return -1, fmt.Errorf("error encoding BINT %s", ErrEncode.Error())
	}
	bw.WriteString("e")
	wlen += n

	return wlen, nil
}

func (o *BInt) Decode(br *bufio.Reader, node *BNode) {
	node.type_ = BINT
	br.ReadByte()
	str, _ := br.ReadString('e')
	str = strings.Trim(str, "e")
	//fmt.Println(str)
	num, _ := strconv.Atoi(str)
	*o = BInt(num)
	node.data = o
}

type BStr string

func (o *BStr) Encode(writer io.Writer) (int, error) {
	bw, ok := writer.(*bufio.Writer)
	if !ok {
		return -1, ErrTyp
	}
	str := string(*o)
	length := len(str)
	lenstr := strconv.Itoa(length)
	n1, err := bw.WriteString(lenstr)
	if err != nil {
		return -1, fmt.Errorf("error encoding BSTR %s", ErrEncode.Error())
	}
	n2, err := bw.WriteString(":" + str)
	if err != nil {
		return -1, fmt.Errorf("error encoding BSTR %s", ErrEncode.Error())
	}
	wlen := n1 + n2
	return wlen, nil
}

func (o *BStr) Decode(br *bufio.Reader, node *BNode) {
	node.type_ = BSTR
	str, _ := br.ReadString(':')
	str = strings.Trim(str, ":")
	length, _ := strconv.Atoi(str)
	buf := make([]byte, length)
	_, err := io.ReadAtLeast(br, buf, length)
	if err != nil {
		return
	}
	*o = BStr(buf)
	//fmt.Printf("%s:%s\n", str, val)
	node.data = o
}

func (o *BStr) Type() Btype {
	return BSTR
}

type BDict map[string]BObject

func (o *BDict) Encode(writer io.Writer) (int, error) {
	bw, ok := writer.(*bufio.Writer)
	if !ok {
		return -1, ErrTyp
	}
	mp := map[string]BObject(*o)
	wlen := 0
	bw.WriteString("d")
	var elemLen int
	var err error
	var bkey BStr
	for key, val := range OrderIter(mp) {
		bkey = BStr(key)
		elemLen, err = bkey.Encode(bw)
		if err != nil {
			return -1, ErrEncode
		}
		wlen += elemLen
		elemLen, err = val.Encode(bw)
		if err != nil {
			return -1, ErrEncode
		}
		wlen += elemLen
	}
	bw.WriteString("e")
	return wlen + 2, nil

}

func (o *BDict) Decode(br *bufio.Reader, node *BNode) {
	node.type_ = BDICT
	br.ReadByte()
	for t, err := br.Peek(1); err != io.EOF && t[0] != 'e'; t, err = br.Peek(1) {
		str, _ := br.ReadString(':')
		str = strings.Trim(str, ":")
		length, _ := strconv.Atoi(str)
		buf := make([]byte, length)
		_, err := io.ReadAtLeast(br, buf, length)
		if err != nil {
			return
		}
		key := string(buf)
		elem, err := BenDecode(br)
		if err != nil {
			return
		}
		//fmt.Printf("%s:%v\n", key, *elem)
		(*o)[key] = elem.data
	}
	br.ReadByte()
	node.data = o
}

func (o *BDict) Type() Btype {
	return BDICT
}

type BList []BObject

func (o *BList) Encode(writer io.Writer) (int, error) {
	bw, ok := writer.(*bufio.Writer)
	if !ok {
		return -1, ErrTyp
	}
	wlen := 0
	bw.WriteString("l")
	var elemLen int
	var err error
	for _, v := range *o {
		elemLen, err = v.Encode(bw)
		if err != nil {
			return -1, ErrEncode
		}
		wlen += elemLen
	}
	bw.WriteString("e")
	return wlen + 2, nil

}

func (o *BList) Decode(br *bufio.Reader, node *BNode) {
	node.type_ = BLIST
	br.ReadByte()
	for t, _ := br.Peek(1); t[0] != 'e'; t, _ = br.Peek(1) {
		elem, err := BenDecode(br)
		if err != nil {
			return
		}
		*o = append(*o, elem.data)
	}
	br.ReadByte()
	node.data = o
}

func (o *BList) Type() Btype {
	return BLIST
}

type BNode struct {
	type_ Btype
	data  BObject
}

func Encode(o *BNode, writer io.Writer) (int, error) {
	bw, ok := writer.(*bufio.Writer)
	if !ok {
		return -1, ErrTyp
	}
	wlen := 0
	length, err := o.data.Encode(bw)
	if err != nil {
		return -1, ErrEncode
	}
	wlen += length
	writeline(bw, &wlen)
	err = bw.Flush()
	if err != nil {
		return -1, fmt.Errorf("%s %s", ErrEncode.Error(), " writer flush ")
	}
	return wlen, nil
}

func DecodeFromFile(name string) ([]*BNode, error) {
	fd, err := os.OpenFile("./test.file", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	fmt.Println("open file successfully")
	reader := bufio.NewReader(fd)
	return recursiveDecode(reader)
}
func DecodeFromString(str string) ([]*BNode, error) {
	reader := bufio.NewReader(strings.NewReader(str))
	return recursiveDecode(reader)
}

func recursiveDecode(reader *bufio.Reader) ([]*BNode, error) {
	nodes := make([]*BNode, 0)
	for {
		parse, err := BenDecode(reader)
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
func BenDecode(r io.Reader) (*BNode, error) {
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
	case b[0] >= '0' && b[0] <= '9':
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
			val := make(BDict)
			val.Decode(br, node)
		}
	case b[0] == '\n' || b[0] == ' ':
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

func OrderIter(mp BDict) iter.Seq2[string, BObject] {
	keys := make([]string, len(mp))
	i := 0
	for k, _ := range mp {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	return func(yield func(string2 string, object BObject) bool) {
		for _, k := range keys {
			if !yield(k, mp[k]) {
				return
			}
		}
	}

}
func writeline(w *bufio.Writer, wlen *int) {
	*wlen++
	w.WriteByte('\n')
}
func writeSpace(w *bufio.Writer, wlen *int) {
	*wlen++
	w.WriteByte(' ')
}
func PrintBobj(b BObject, tab string) {
	switch v := b.(type) {
	case *BInt:
		{
			fmt.Println(tab+"Bint : ", *v)
		}
	case *BStr:
		{
			fmt.Println(tab+"BStr : ", *v)
		}
	case *BList:
		{
			fmt.Println(tab + "BList : ")
			for _, v := range *v {
				PrintBobj(v, tab+"  ")
			}
		}
	case *BDict:
		{
			fmt.Println(tab + "BDict : ")
			for k, v := range *v {
				fmt.Printf(tab+"  key:%s : ", k)
				PrintBobj(v, tab+"  ")
			}
		}
	}
}
