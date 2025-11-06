package bmatch

import (
	"regexp"
	"strings"
	"testing"

	"github.com/cvilsmeier/bmatch/internal"
)

func TestMatcher(t *testing.T) {
	explain := func(expr string) string {
		plan, err := Explain(expr)
		if err != nil {
			return "err: " + err.Error()
		}
		return plan
	}
	t.Run("matchEverything", func(t *testing.T) {
		is := internal.Assert(t)
		for _, expr := range []string{"", "//"} {
			is.Eq("//", explain(expr))
			m := MustCompile(expr)
			is.True(m.Match(""))
			is.True(m.Match("a"))
			is.True(m.Match("ababab"))
			is.True(m.Match("aa"))
			is.True(m.Match("bb"))
			is.True(m.Match("aabb"))
			is.True(m.Match("bb  aa"))
			is.True(m.Match("bb  aabb  aa"))
		}
	})
	t.Run("matchNothing", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "NOT //"
		is.Eq("NOT[//]", explain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.False(m.Match("a"))
	})
	t.Run("matchOnlyEmptyString", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "/^$/"
		is.Eq("/^$/", explain(expr))
		m := MustCompile(expr)
		is.True(m.Match(""))
		is.False(m.Match("a"))
	})
	t.Run("matchOnlyNonEmptyString", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "NOT /^$/"
		is.Eq("NOT[/^$/]", explain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.True(m.Match("a"))
	})
	t.Run("andOperator", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "/aa/ AND /bb/"
		is.Eq("AND[/aa/,/bb/]", explain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.False(m.Match("a"))
		is.False(m.Match("ababab"))
		is.False(m.Match("aa"))
		is.False(m.Match("bb"))
		is.True(m.Match("aabb"))
		is.True(m.Match("bb  aa"))
		is.True(m.Match("bb  aabb  aa"))
	})
	t.Run("andOr", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "/foo/ AND /bar/ OR /baz/"
		is.Eq("OR[AND[/foo/,/bar/],/baz/]", explain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.False(m.Match("foo"))
		is.False(m.Match("bar"))
		is.True(m.Match("baz"))
		is.True(m.Match("foobar"))
		is.True(m.Match("foobaz"))
		is.True(m.Match("barbaz"))
		is.True(m.Match("foobarbaz"))
		is.True(m.Match("barfoobaz"))
		is.True(m.Match("bazbarfoo"))
	})
	t.Run("andOrGroup", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "/foo/ AND (/bar/ OR /baz/)"
		is.Eq("AND[/foo/,OR[/bar/,/baz/]]", explain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.False(m.Match("foo"))
		is.False(m.Match("bar"))
		is.False(m.Match("baz"))
		is.True(m.Match("foobar"))
		is.True(m.Match("foobaz"))
		is.False(m.Match("barbaz"))
		is.True(m.Match("foobarbaz"))
		is.True(m.Match("barfoobaz"))
		is.True(m.Match("bazbarfoo"))
	})
	t.Run("withEscaping", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "/^\\/home/ AND NOT /^\\/home\\/tmp/"
		is.Eq("AND[/^/home/,NOT[/^/home/tmp/]]", explain(expr))
		m := MustCompile(expr)
		is.True(m.Match("/home/joe/bin"))
		is.False(m.Match("/home/tmp/bin"))
		is.True(m.Match("/home/joe/tmp"))
		is.False(m.Match("/opt/home/joe/tmp"))
	})
	t.Run("readme", func(t *testing.T) {
		is := internal.Assert(t)
		expr := "/DEBUG/ OR ( /TRACE/ AND NOT /(?i)TRACE.*sql/ )"
		is.Eq("OR[/DEBUG/,AND[/TRACE/,NOT[/(?i)TRACE.*sql/]]]", explain(expr))
		m := MustCompile(expr)
		is.True(m.Match("10:00 DEBUG will poll now"))
		is.True(m.Match("10:01 DEBUG polling error: no route to host"))
		is.True(m.Match("10:02 TRACE(poll) connecting 10.0.0.21"))
		is.False(m.Match("10:03 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.False(m.Match("10:04 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.True(m.Match("10:05 TRACE(http) POST /contactForm from 174.161.32.109"))
		is.False(m.Match("10:06 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.True(m.Match("11:00 DEBUG will poll now"))
	})
}

const randomText = "Cause dried no solid no an small so still widen. Ten weather evident smiling bed against she examine its. Rendered far opinions two yet moderate sex striking. Sufficient motionless compliment by stimulated assistance at. Convinced resolving extensive agreeable in it on as remainder. Cordially say affection met who propriety him. Are man she towards private weather pleased. In more part he lose need so want rank no. At bringing or he sensible pleasure. Prevent he parlors do waiting be females an message society."
const randomTextWithOscarPeterson = "Cause dried no solid no Peterson an small so still widen. Ten weather evident smiling bed against she examine its. Rendered far opinions two yet moderate sex striking. Sufficient motionless compliment by stimulated assistance at. Convinced resolving extensive agreeable in it on as remainder. Cordially say affection met who propriety him. Are man she towards private weather pleased. In more part he lose need so want rank no. At bringing or Oscar he sensible pleasure. Prevent he parlors do waiting Oscar Peterson be females an message society."

func BenchmarkStringsContains(b *testing.B) {
	const expr = "Oscar Peterson"
	for i := 0; i < b.N; i++ {
		if strings.Contains(randomText, expr) {
			b.Fatalf("must not match")
		}
		if !strings.Contains(randomTextWithOscarPeterson, expr) {
			b.Fatalf("must match")
		}
	}
}

func BenchmarkRegexpMatch(b *testing.B) {
	const expr = "Oscar Peterson"
	rex := regexp.MustCompile(expr)
	for i := 0; i < b.N; i++ {
		if rex.MatchString(randomText) {
			b.Fatalf("must not match")
		}
		if !rex.MatchString(randomTextWithOscarPeterson) {
			b.Fatalf("must match")
		}
	}
}

func BenchmarkBmatchMatch(b *testing.B) {
	const expr = "/Oscar Peterson/"
	matcher := MustCompile(expr)
	for i := 0; i < b.N; i++ {
		if matcher.Match(randomText) {
			b.Fatalf("must not match")
		}
		if !matcher.Match(randomTextWithOscarPeterson) {
			b.Fatalf("must match")
		}
	}
}
