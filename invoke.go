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
		for {
			// check to see if we already have a value for this input type and use it
			// if we do. ignore err here because after validate() this can no longer
			// return an error.
			val, ok, name, _ = r.provide(t.In(i))
			if ok && val.IsValid() && val.Kind() != reflect.Func {
				break
			}
			if val.Kind() != reflect.Func {
				return fmt.Errorf("missing provider object for %v", name)
			}
			// call the provider function for this input argument, since it has not
			// yet been resolved.
			if err := r.callfn(val, callmap); err != nil {
				return err
			}
		}

		// Either the argument is a struct and we have a pointer, or the argument
		// is a pointer and we have a struct, so we need to normalize our val.
		if val.Kind() == reflect.Ptr && t.In(i).Kind() == reflect.Struct {
			// turn a pointer into a struct value by dereferencing
			val = val.Elem()
		} else if val.Kind() == reflect.Struct && t.In(i).Kind() == reflect.Ptr {
			// to get a pointer from a struct value, we need to allocate a new pointer
			// object and set the struct as the address of the pointer. We do that
			// with this little bit of reflection magic:
			v := reflect.New(val.Type())
			v.Elem().Set(val)
			val = v
		}
		ins[i] = val
	}

	// call the provider function to generate the provided values.
	outs := fn.Call(ins)
	for _, out := range outs {
		// handle error returns
		if iserr(out.Type()) && !out.IsNil() {
			return out.Interface().(error)
		}

		// identify what provided object this might be (if any)
		c := r.concrete(out.Type())
		if c == nil {
			continue
		}
		if name := nameof(c); name != "" {
			r.valmap[name] = out
		}
	}
	return nil
}
