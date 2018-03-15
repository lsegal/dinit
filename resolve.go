package dinit

import (
	"fmt"
	"reflect"
)

type resolver struct {
	// the map of type names to their Type object
	typmap map[string]reflect.Type

	// the map of provided values. if the value is a function, it can be invoked
	// to return an object of the named type.
	valmap map[string]reflect.Value

	// the list of functions to sort based on the depmap
	fns []reflect.Value
}

// init initializes the valmap and fns list. Concrete values are separated from
// provider functions and will be used later during invoke().
func (r *resolver) init(vals []interface{}) error {
	r.valmap = map[string]reflect.Value{}
	for _, val := range vals {
		v := reflect.ValueOf(val)
		switch v.Kind() {
		case reflect.Func:
			r.fns = append(r.fns, v)
		case reflect.Struct, reflect.Ptr:
			r.valmap[nameof(v.Type())] = v
		default:
			return fmt.Errorf("unsupported provider type: %v", val)
		}
	}
	return nil
}

// resolve inspects the list of concrete values and provider functions to
// determine the dependency map of the initialization.
func (r *resolver) resolve() error {
	r.buildTypeMap()
	return r.fillValMap()
}

// buildTypeMap collects all known type names from all concrete values, as well
// as input and output function argument types.
func (r *resolver) buildTypeMap() {
	r.typmap = map[string]reflect.Type{}
	for _, fn := range r.fns {
		fnt := fn.Type()
		for i := 0; i < fnt.NumIn(); i++ {
			r.addType(fnt.In(i))
		}
		for i := 0; i < fnt.NumOut(); i++ {
			r.addType(fnt.Out(i))
		}
	}
	for _, v := range r.valmap {
		r.addType(v.Type())
	}
}

// addType adds a type to the known types list.
func (r *resolver) addType(t reflect.Type) {
	t = elemof(t)
	if t.Kind() == reflect.Interface {
		return
	}
	if name := nameof(t); name != "" {
		r.typmap[name] = t
	}
}

// fillValMap fills the valmap with the remaining provider functions for the
// matching types.
func (r *resolver) fillValMap() error {
	for _, fn := range r.fns {
		fnt := fn.Type()
		for i := 0; i < fnt.NumOut(); i++ {
			c := r.concrete(fnt.Out(i))
			if c == nil {
				continue
			}
			name := nameof(c)
			if name == "" {
				continue
			}
			if _, ok := r.valmap[name]; ok {
				continue // skip if we already have a value
			}
			r.valmap[name] = fn
		}
	}
	return nil
}

// concrete determines the concrete name of a type. In other words, if the type
// is an interface, it determines the closest matching known concrete type
// that implements said interface. When a match is found, the return value of
// this function will always be a struct value (not pointer) for consistency. If
// no match is found, this function returns nil.
func (r *resolver) concrete(t reflect.Type) reflect.Type {
	t = elemof(t)
	name := nameof(t)
	if name == "" {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
		return t
	case reflect.Interface:
		for _, v := range r.typmap {
			if reflect.PtrTo(v).Implements(t) {
				return v
			}
		}
	}
	return nil
}

// provide returns the provide value for a given struct/interface type. If the
// value is a function, it should be invoked to return the value.
func (r *resolver) provide(t reflect.Type) (v reflect.Value, ok bool, name string, err error) {
	c := r.concrete(t)
	if c == nil {
		err = fmt.Errorf("no injectable value for type %v", t)
		return
	}
	name = nameof(c)
	if name == "" {
		err = fmt.Errorf("unknown type %v", t)
		return
	}
	v, ok = r.valmap[name]
	return
}

// validate checks to see if any provider functions will create a call cycle
// when trying to initialize objects or if any arguments are unknown.
func (r *resolver) validate(fn reflect.Value, m map[reflect.Value]int) error {
	if !fn.IsValid() {
		return nil
	}
	fnt := fn.Type()

	if m == nil {
		m = map[reflect.Value]int{}
	}
	m[fn]++
	defer func() { m[fn]-- }()
	if m[fn] >= 2 {
		return fmt.Errorf("cycle detected in %+v", fnt)
	}

	for i := 0; i < fnt.NumIn(); i++ {
		pfn, ok, _, err := r.provide(fnt.In(i))
		if err != nil {
			return err
		}
		if !ok || pfn.Kind() != reflect.Func {
			continue
		}
		if err := r.validate(pfn, m); err != nil {
			return err
		}
	}
	return nil
}
