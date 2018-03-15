package dinit

import (
	"reflect"
	"strings"
)

// elemof gets the value-type of a given type that might be a pointer.
func elemof(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// nameof returns the fully qualified type name of a Type object.
func nameof(t reflect.Type) string {
	t = elemof(t)
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Interface {
		return ""
	} else if iserr(t) {
		return ""
	}
	return strings.Join([]string{t.PkgPath(), t.Name()}, ".")
}

// iserr returns true if the type conforms to the error interface.
func iserr(t reflect.Type) bool {
	if t.Kind() == reflect.Interface {
		return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
	}
	return false
}
