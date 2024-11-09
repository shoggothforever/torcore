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
	l := reflect.MakeSlice(rcValue.Elem().Type(), len(list), len(list))
	elem.Set(l)
	for k, v := range list {
		switch v.Type() {
		case BINT:
			elem.Index(k).SetInt(int64(*v.(*BInt)))
		case BSTR:
			elem.Index(k).SetString(string(*v.(*BStr)))
		case BLIST:
			if reflect.TypeOf(v).Elem().Kind() != reflect.Slice {
				return fmt.Errorf("it's invalid to marshal map elem into type other than slice")
			}
			ln := reflect.New(elem.Type())
			err := unmarshalList(ln, *v.(*BList))
			if err != nil {
				return err
			}
			elem.Index(k).Set(ln.Elem())
		case BDICT:
			if reflect.TypeOf(v).Elem().Kind() != reflect.Struct {
				return fmt.Errorf("it's invalid to marshal map elem into type other than struct")
			}
			ln := reflect.New(reflect.TypeOf(v).Elem())
			err := unmarshalDict(ln, *v.(*BDict))
			if err != nil {
				return err
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
	for i, n := 0, elem.NumField(); i < n; i++ {
		fv := elem.Field(i)
		if !fv.CanSet() {
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
				ln := reflect.Value{}
				ln = reflect.New(fv.Type())
				if ft.Type.Kind() != reflect.Slice {
					if ft.Type.Kind() != reflect.Ptr || ft.Type.Elem().Kind() != reflect.Struct {
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

			default:
				return ErrMarshal
			}

		}

	}

	return nil

}
