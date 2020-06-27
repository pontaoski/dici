package xattr

import (
	"reflect"

	"github.com/pkg/xattr"
)

// ApplyXattrs applies xattrs from a struct annotated with xattr tags onto a file
func ApplyXattrs(file string, v interface{}) (err error) {
	t := reflect.TypeOf(v)
	d := reflect.ValueOf(v)
	for i := 0; i < t.NumField(); i++ {
		if !(t.Field(i).Type.Kind() == reflect.String) {
			continue
		}
		val := d.Field(i).String()

		name, ok := t.Field(i).Tag.Lookup("xattr")
		if !ok {
			continue
		}
		if err != nil {
			continue
		}
		err = addXattr(file, "user.dici:"+name, val)
	}
	return
}

// ReadXattrs reads xattrs from a file into a struct annotated with xattr tags
func ReadXattrs(file string, v interface{}) (err error) {
	d := reflect.ValueOf(v).Elem()
	t := d.Type()
	for i := 0; i < t.NumField(); i++ {
		if !(t.Field(i).Type.Kind() == reflect.String) {
			continue
		}

		var data string
		name, ok := t.Field(i).Tag.Lookup("xattr")
		if !ok {
			continue
		}
		if err != nil {
			continue
		}
		data, err = readXattr(file, "user.dici:"+name)

		d.Field(i).Set(reflect.ValueOf(data))
	}
	d = reflect.ValueOf(v)
	return
}

func addXattr(file string, key string, value string) error {
	return xattr.Set(file, key, []byte(value))
}

func readXattr(file string, key string) (string, error) {
	data, err := xattr.Get(file, key)
	return string(data), err
}
