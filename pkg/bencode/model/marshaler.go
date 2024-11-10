package model

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// r: bencode字符串或者文件的读取流
// receiver: 接收者,可以是int|string|map[string]*Bobject|[]*Bobject
func Unmarshal(r io.Reader, receiver interface{}) error {
	o, err := BenDecode(r)
	if err != nil {
		return err
	}
	//PrintBobj(o.data, "")
	rcValue := reflect.ValueOf(receiver)
	if rcValue.Kind() != reflect.Ptr {
		return ErrMarshal
	}
	switch o.type_ {
	case BLIST:
		list, ok := o.data.(*BList)
		if !ok {
			return ErrMarshal
		}
		err = unmarshalList(rcValue, *list)
		if err != nil {
			return err
		}
	case BDICT:
		dict, ok := o.data.(*BDict)
		if !ok {
			return ErrMarshal
		}
		err = unmarshalDict(rcValue, *dict)
		if err != nil {
			return err
		}
	default:
		return ErrMarshal

	}

	return nil
}

// reflect.Value: 用于赋值的Bobject对象的value
// receiver: 接收者的指针
func unmarshalList(rcValue reflect.Value, list BList) error {
	if rcValue.Kind() != reflect.Ptr || rcValue.Elem().Kind() != reflect.Slice {
		return ErrMarshal
	}
	if len(list) == 0 {
		return nil
	}
	elem := rcValue.Elem()
	elemSlice := reflect.MakeSlice(elem.Type(), len(list), len(list))
	elem.Set(elemSlice)

	fmt.Println(elem)
	for k, v := range list {
		fmt.Println("Index:", k, "Type:", reflect.TypeOf(v))
		switch o := v.(type) {
		case *BInt:
			if elem.Index(k).CanSet() {
				elem.Index(k).Set(reflect.ValueOf(int(*o)))
			} else {
				fmt.Println(elem.Index(k).Type())
				fmt.Println("Cannot set BStr")
			}
		case *BStr:
			if elem.Index(k).CanSet() {
				elem.Index(k).Set(reflect.ValueOf(string(*o)))
			} else {
				fmt.Println(elem.Index(k).Type())
				fmt.Println("Cannot set BStr")
			}
		case *BList:
			if reflect.TypeOf(v).Elem().Kind() != reflect.Slice {
				return fmt.Errorf("it's invalid to marshal map elem into type other than slice")
			}
			ln := reflect.New(elem.Index(k).Type())
			err := unmarshalList(ln, *o)
			if err != nil {
				return err
			}
			elem.Index(k).Set(ln.Elem())
		case *BDict:
			if elem.Index(k).Kind() != reflect.Ptr && elem.Index(k).Kind() != reflect.Struct {
				return fmt.Errorf("it's invalid to marshal map elem into type other than struct %s", elem.Index(k).Kind())
			} else if elem.Index(k).Kind() == reflect.Ptr && elem.Index(k).Elem().Kind() != reflect.Struct {
				return fmt.Errorf("it's invalid to marshal map elem into type other than struct %s", elem.Index(k).Elem().Kind())
			}
			ln := reflect.New(elem.Index(k).Type())
			err := unmarshalDict(ln, *o)
			if err != nil {
				return fmt.Errorf("unmarshalDict %s", err.Error())
			}
			elem.Index(k).Set(ln.Elem())

		default:
			return ErrMarshal

		}
	}

	return nil
}

func unmarshalDict(rcValue reflect.Value, dict BDict) error {
	if rcValue.Kind() != reflect.Ptr || (rcValue.Elem().Kind() != reflect.Struct && rcValue.Elem().Kind() != reflect.Map) {
		return ErrMarshal
	}
	elem := rcValue.Elem()
	tp := elem.Type()
	if elem.Kind() == reflect.Struct {
		for i, n := 0, len(dict); i < n; i++ {
			fv := elem.Field(i)
			if !fv.CanSet() {
				fmt.Println("can not set")
				continue
			}
			ft := tp.Field(i)
			tag := ft.Tag.Get(Btag)
			if len(tag) == 0 {
				tag = strings.ToLower(ft.Name)
			}
			if v, ok := dict[tag]; ok {
				switch v.Type() {
				case BINT:
					if ft.Type.Kind() < reflect.Int || ft.Type.Kind() > reflect.Int64 {
						break
					}
					fv.SetInt(int64(*v.(*BInt)))
				case BSTR:
					if ft.Type.Kind() != reflect.String {
						break
					}
					fv.SetString(string(*v.(*BStr)))
				case BLIST:
					ln := reflect.New(fv.Type())
					if ft.Type.Kind() != reflect.Slice {
						if ft.Type.Kind() != reflect.Ptr || ft.Type.Elem().Kind() != reflect.Slice {
							break
						} else {
							ln = reflect.New(fv.Elem().Type())
						}
					}
					err := unmarshalList(ln, *v.(*BList))
					if err != nil {
						return err
					}
					fv.Set(ln.Elem())
				case BDICT:
					if ft.Type.Kind() != reflect.Struct {
						break
					}
					dp := reflect.New(ft.Type)
					dict := *v.(*BDict)
					err := unmarshalDict(dp, dict)
					if err != nil {
						break
					}
					fv.Set(dp.Elem())
				default:
					return ErrMarshal
				}

			}

		}
	} else if elem.Kind() == reflect.Map {
		keys := elem.MapKeys()
		for _, key := range keys {
			// 获取 map 中的值
			value := elem.MapIndex(key)
			fmt.Printf("Key: %v, Value: %v\n", key.Interface(), value.Interface())
		}
	}

	return nil

}

func Marshal(w io.Writer, v interface{}) int {

	p := reflect.ValueOf(v)
	if p.Kind() == reflect.Ptr {
		p = p.Elem()
	}

	return MarshalValue(w, p)
}
func MarshalValue(w io.Writer, v reflect.Value) int {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	l := 0
	n := 0
	switch v.Kind() {
	case reflect.Struct:
		l += MarshalDict(bw, v)
	case reflect.Slice:
		fmt.Println("marshal list before")
		l += MarshalList(bw, v)
		fmt.Println("marshal list after , write ", l)
	case reflect.Int:
		bInt := BInt(v.Int())
		n, _ = bInt.Encode(bw)
		l += n
	case reflect.String:
		//fmt.Println("marshal ", v.String())
		bStr := BStr(v.String())
		n, _ = bStr.Encode(bw)
		l += n
	}
	err := bw.Flush()
	if err != nil {
		return 0
	}
	return l
}
func MarshalList(w io.Writer, v reflect.Value) int {
	l := 2
	_, err := w.Write([]byte{'l'})
	if err != nil {
		return -1
	}
	if v.Len() == 0 {
		fmt.Println("get empty slice")
	}
	for i := 0; i < v.Len(); i++ {
		l += MarshalValue(w, v.Index(i))
	}
	_, err = w.Write([]byte{'e'})
	if err != nil {
		return -1
	}
	return l

}

func MarshalDict(w io.Writer, v reflect.Value) int {
	l := 2
	_, err := w.Write([]byte{'d'})
	if err != nil {
		return -1
	}
	for i := 0; i < v.NumField(); i++ {
		ft := v.Type().Field(i)
		fv := v.Field(i)
		ben := ft.Tag.Get(Btag)
		if len(ben) == 0 {
			ben = strings.ToLower(ft.Name)
		}
		str := BStr(ben)
		n, err := str.Encode(w)
		if err != nil {
			return -1
		}
		l += n
		l += MarshalValue(w, fv)
	}
	_, err = w.Write([]byte{'e'})
	if err != nil {
		return -1
	}

	return l

}
