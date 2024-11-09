package model

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// r: bencode字符串或者文件的读取流
// receiver: 接收者,可以是int|string|map[string]*Bobject|[]*Bobject
func Unmarshal(r io.Reader, receiver interface{}) error {
	o, err := BenParse(r)
	if err != nil {
		return err
	}
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
		//
	}

	return nil
}

func Marshal(w io.Writer, v interface{}) error {
	//bw, ok := w.(*bufio.Writer)
	//if !ok {
	//	bw = bufio.NewWriter(w)
	//}
	//p := reflect.ValueOf(v)
	//if p.Kind() != reflect.Ptr {
	//	return ErrMarshal
	//}

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
	elemSlice := make(BList, len(list), len(list))
	for k := range list {
		elemSlice[k] = list[k]
	}
	l := reflect.ValueOf(elemSlice)
	elem.Set(l)
	fmt.Println(elem)
	for k, v := range list {
		fmt.Println("Index:", k, "Type:", reflect.TypeOf(v))
		switch o := v.(type) {
		case *BStr, *BInt:
			if elem.Index(k).CanSet() {
				elem.Index(k).Set(reflect.ValueOf(o))
			} else {
				fmt.Println(elem.Index(k).Type())
				fmt.Println("Cannot set BStr")
			}
		case *BList:
			if reflect.TypeOf(v).Elem().Kind() != reflect.Slice {
				return fmt.Errorf("it's invalid to marshal map elem into type other than slice")
			}
			ln := reflect.New(elem.Type())
			err := unmarshalList(ln, *o)
			if err != nil {
				return err
			}
			elem.Index(k).Set(ln.Elem())
		case *BDict:
			if reflect.TypeOf(v).Elem().Kind() != reflect.Struct {
				return fmt.Errorf("it's invalid to marshal map elem into type other than struct %s", reflect.TypeOf(v).Elem().Kind())
			}
			ln := reflect.New(reflect.TypeOf(v).Elem())
			err := unmarshalDict(ln, *o)
			if err != nil {
				return fmt.Errorf("unmarshalDict %s", err.Error())
			}
			elem.Index(k).Set(ln)

		default:
			return ErrMarshal

		}
	}

	return nil
}

func unmarshalDict(rcValue reflect.Value, dict BDict) error {
	if rcValue.Kind() != reflect.Ptr || rcValue.Elem().Kind() != reflect.Struct {
		return ErrMarshal
	}
	elem := rcValue.Elem()
	tp := elem.Type()
	for i, n := 0, len(dict); i < n; i++ {
		fv := elem.Field(i)
		if !fv.CanSet() {
			fmt.Println("can not set")
			continue
		}
		ft := tp.Field(i)
		tag := ft.Tag.Get("bencode")
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

	return nil

}
