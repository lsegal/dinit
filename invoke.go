package dinit

import (
	"fmt"
	"reflect"
)

// invoke calls all provider functions in the best order possible, using the
// generated output values as inputs to dependent providers.
func (r *resolver) invoke() error {
	callmap := map[reflect.Value]bool{}
	for _, fn := range r.fns {
		if callmap[fn] {
			continue
		}
		if err := r.callfn(fn, callmap); err != nil {
			return err
		}
	}
	return nil
}

// callfn calls a provider function with the correct input arguments matched
// against the current set of resolved values. If a function requires an input
// value that has not yet been generated, callfn will look up the provider
// function for that type and generate the value recursively.
func (r *resolver) callfn(fn reflect.Value, callmap map[reflect.Value]bool) error {
	callmap[fn] = true
	t := fn.Type()

	// this will be the map of dependent types that need producing
	ins := make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		var val reflect.Value
		var ok bool
		var name string
		var err error
		for {
			val, ok, name, err = r.provide(t.In(i))
			if err != nil {
				return err
			}
			if ok && val.IsValid() && val.Kind() != reflect.Func {
				break
			}
			if val.Kind() != reflect.Func {
				return fmt.Errorf("missing provider object for %v", name)
			}
			if err := r.callfn(val, callmap); err != nil {
				return err
			}
		}

		if val.Kind() == reflect.Ptr && t.In(i).Kind() == reflect.Struct {
			val = val.Elem()
		} else if val.Kind() == reflect.Struct && t.In(i).Kind() == reflect.Ptr {
			v := reflect.New(val.Type())
			v.Elem().Set(val)
			val = v
		}
		ins[i] = val
	}

	outs := fn.Call(ins)
	for _, out := range outs {
		if iserr(out.Type()) && !out.IsNil() {
			return out.Interface().(error)
		}

		c := r.concrete(out.Type())
		if c == nil {
			continue
		}
		name := nameof(c)
		if name == "" {
			continue
		}
		r.valmap[name] = out
	}
	return nil
}
