package internal

import (
	"testing"
)

// An Asserter provides assertion methods for unit tests.
type Asserter struct {
	t testing.TB
}

func Assert(t testing.TB) *Asserter {
	return &Asserter{t}
}

func (a *Asserter) Failf(format string, args ...any) {
	a.t.Helper()
	a.t.Fatalf(format, args...)
}

func (a *Asserter) True(c bool) {
	a.t.Helper()
	if !c {
		a.t.Fatalf("\nwant true\nhave %v", c)
	}
}

func (a *Asserter) False(c bool) {
	a.t.Helper()
	if c {
		a.t.Fatalf("\nwant false\nhave %v", c)
	}
}

func (a *Asserter) Eq(want, have any) {
	a.t.Helper()
	if want != have {
		a.t.Fatalf("\nwant %#v\nhave %#v", want, have)
	}
}
