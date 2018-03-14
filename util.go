package dinit

import (
	"reflect"
	"strings"
)

func elemof(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func nameof(t reflect.Type) string {
	t = elemof(t)
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Interface {
		return ""
	} else if iserr(t) {
		return ""
	}
	return strings.Join([]string{t.PkgPath(), t.Name()}, ".")
}

func iserr(t reflect.Type) bool {
	if t.Kind() == reflect.Interface {
		return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
	}
	return false
}
