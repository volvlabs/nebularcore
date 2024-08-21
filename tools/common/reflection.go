package common

import "reflect"

func GetStructName(s interface{}) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
