package bmatch

import (
	"testing"
)

type asserter struct {
	t testing.TB
}

func assert(t testing.TB) *asserter {
	return &asserter{t}
}

func (a *asserter) failf(format string, args ...any) {
	a.t.Helper()
	a.t.Fatalf(format, args...)
}

func (a *asserter) true(c bool) {
	a.t.Helper()
	if !c {
		a.t.Fatalf("\nwant true\nhave %v", c)
	}
}

func (a *asserter) false(c bool) {
	a.t.Helper()
	if c {
		a.t.Fatalf("\nwant false\nhave %v", c)
	}
}

func (a *asserter) eq(want, have any) {
	a.t.Helper()
	if want != have {
		a.t.Fatalf("\nwant %#v\nhave %#v", want, have)
	}
}
