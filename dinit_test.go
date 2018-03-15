package dinit_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lsegal/dinit"
	"github.com/stretchr/testify/assert"
)

var (
	out []string
)

type a struct {
	b b
	c c
}

type b struct {
	c cer
}

type c struct {
	val string
}

type d struct {
	val string
}

func (c c) C() string {
	return c.val
}

type cer interface {
	C() string
}

func newA(b *b, c c) *a {
	out = append(out, "INIT A")
	return &a{b: *b, c: c}
}

func newB(c cer, d d) (b, error) {
	if c.C() == "" {
		return b{}, errors.New("invalid argument")
	}
	out = append(out, "INIT B: "+d.val)
	return b{c: c}, nil
}

func newC() *c {
	out = append(out, "INIT C")
	return &c{val: "hello"}
}

type e struct {
}

type f struct {
}

func newE(f1, f2 f) e {
	out = append(out, "E")
	return e{}
}

func newF(e e) f {
	out = append(out, "F")
	return f{}
}

func TestInit(t *testing.T) {
	out = []string{}
	err := dinit.Init(newA, newB, newC, d{"test"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"INIT C", "INIT B: test", "INIT A"}, out)
}

func TestInit_Error(t *testing.T) {
	out = []string{}
	err := dinit.Init(newB, c{""}, d{"test"})
	assert.EqualError(t, err, "invalid argument")
	assert.Equal(t, []string{}, out)
}

func TestInit_Cycle(t *testing.T) {
	out = []string{}
	err := dinit.Init(newE, newF)
	assert.EqualError(t, err, "cycle detected in func(didi_test.f, didi_test.f) didi_test.e")
	assert.Equal(t, []string{}, out)
}

func ExampleInit() {
	type C struct{ Value string }
	type B struct{ *C }
	type A struct {
		*B
		*C
	}

	// Standard Go-style New* initializer functions
	NewA := func(b *B, c *C) *A { return &A{b, c} }
	NewB := func(c *C) *B { return &B{c} }

	// Use this one to inspect our produced A value
	finalizer := func(a *A) { fmt.Println(a.C.Value, a.B.C.Value) }

	// Pass in a static value for C so we can control the string Value.
	// DInit is smart enough to use the B value produced by NewB as an argument
	// to NewA.
	dinit.Init(finalizer, NewA, NewB, &C{"Winner"})

	// Output: Winner Winner
}
