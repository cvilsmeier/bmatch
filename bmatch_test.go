package bmatch

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/cvilsmeier/bmatch/internal"
)

func TestBmatch(t *testing.T) {
	type input struct {
		text   string
		result bool
	}
	type testcase struct {
		name      string
		expr      string
		planOrErr string
		inputs    []input
	}
	for _, tt := range []testcase{
		{
			"emptyString",
			"",
			"''",
			[]input{
				{"", true},
				{"a", true},
				{"foobar", true},
				{"    foobar  ", true},
			},
		},
		{
			"emptyRegex",
			"//",
			"//",
			[]input{
				{"", true},
				{"a", true},
				{"foobar", true},
				{"    foobar  ", true},
			},
		},
		{
			"simpleString",
			"foo",
			"'foo'",
			[]input{
				{"", false},
				{"foo bar", true},
				{"foo", true},
				{"/foo", true},
				{"foo/", true},
				{"    foo  ", true},
				{"foo foo", true},
				{"foofoo", true},
				{"fo", false},
			},
		},
		{
			"simpleRegex",
			"/fo*bar/",
			"/fo*bar/",
			[]input{
				{"", false},
				{"bar", false},
				{"fbar", true},
				{"fobar", true},
				{"foobar", true},
				{"fooobar", true},
				{"fooolbar", false},
				{"    foobar  ", true},
			},
		},
		{
			"notString",
			"NOT foo",
			"NOT['foo']",
			[]input{
				{"", true},
				{"a", true},
				{"foobar", false},
				{"    foobar  ", false},
				{"    fobar  ", true},
			},
		},
		{
			"notRegex",
			"NOT /foo$/",
			"NOT[/foo$/]",
			[]input{
				{"foo", false},
				{" foo", false},
				{"foo ", true},
				{"foobar", true},
				{"barfoo", false},
			},
		},
		{
			"andExpr",
			"/aa*/ AND bb",
			"AND[/aa*/,'bb']",
			[]input{
				{"", false},
				{"bb", false},
				{"ab", false},
				{"a bb", true},
				{"aa bb", true},
				{"aaa bb", true},
				{"aaabb", true},
				{"aaabab", false},
			},
		},
		{
			"andExprOrExpr",
			"/foo/ AND /bar/ OR /baz/",
			"OR[AND[/foo/,/bar/],/baz/]",
			[]input{
				{"", false},
				{"foo", false},
				{"bar", false},
				{"baz", true},
				{"foobar", true},
				{"foobaz", true},
				{"barbaz", true},
				{"foobarbaz", true},
				{"barfoobaz", true},
				{"bazbarfoo", true},
			},
		},
		{
			"andExprOrExprWithParens",
			"/foo/ AND (/bar/ OR /baz/)",
			"AND[/foo/,OR[/bar/,/baz/]]",
			[]input{
				{"", false},
				{"foo", false},
				{"bar", false},
				{"baz", false},
				{"foobar", true},
				{"foobaz", true},
				{"barbaz", false},
				{"foobarbaz", true},
				{"barfoobaz", true},
				{"bazbarfoo", true},
			},
		},
		{
			"escapingRegex",
			"/^\\/home/ AND NOT /^\\/home\\/tmp/",
			"AND[/^/home/,NOT[/^/home/tmp/]]",
			[]input{
				{"", false},
				{"/home/joe/bin", true},
				{"/home/tmp/bin", false},
				{"/home/joe/tmp", true},
				{"/opt/home/joe/tmp", false},
			},
		},
		{
			"logfileExampleWithoutRegex",
			"DEBUG OR (TRACE AND NOT (TRACE AND SQL))",
			"OR['DEBUG',AND['TRACE',NOT[AND['TRACE','SQL']]]]",
			[]input{
				{"", false},
				{"10:00 DEBUG will poll now", true},
				{"10:01 DEBUG polling error: no route to host", true},
				{"10:02 TRACE(poll) connecting 10.0.0.21", true},
				{"10:03 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?", false},
				{"10:04 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?", false},
				{"10:05 TRACE(http) POST /contactForm from 174.161.32.109", true},
				{"10:06 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?", false},
				{"11:00 DEBUG will poll now", true},
			},
		},
		{
			"logfileExampleWithRegex",
			"/DEBUG/ OR ( /TRACE/ AND NOT /(?i)TRACE.*sql/ )",
			"OR[/DEBUG/,AND[/TRACE/,NOT[/(?i)TRACE.*sql/]]]",
			[]input{
				{"", false},
				{"10:00 DEBUG will poll now", true},
				{"10:01 DEBUG polling error: no route to host", true},
				{"10:02 TRACE(poll) connecting 10.0.0.21", true},
				{"10:03 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?", false},
				{"10:04 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?", false},
				{"10:05 TRACE(http) POST /contactForm from 174.161.32.109", true},
				{"10:06 TRACE(SQL) UPDATE sessions SET lastAccess=? WHERE id=?", false},
				{"11:00 DEBUG will poll now", true},
			},
		},
		{
			"errUnclosedGroup",
			"DEBUG OR (TRACE AND NOT SQL",
			"err: syntax error",
			[]input{},
		},
		{
			"errUnclosedRegex",
			"DEBUG OR /aa",
			"err: unclosed regex in \"aa\"",
			[]input{},
		},
		{
			"errUnbalancedAnd",
			"DEBUG AND",
			"err: syntax error",
			[]input{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := internal.Assert(t)
			if strings.HasPrefix(tt.planOrErr, "err: ") {
				_, err := Explain(tt.expr)
				is.Eq(tt.planOrErr, fmt.Sprintf("err: %s", err))
				_, err = Compile(tt.expr)
				is.Eq(tt.planOrErr, fmt.Sprintf("err: %s", err))
			} else {
				plan, err := Explain(tt.expr)
				is.NoErr(err)
				is.Eq(tt.planOrErr, plan)
				matcher, err := Compile(tt.expr)
				is.NoErr(err)
				for _, input := range tt.inputs {
					is.Eqf(input.result, matcher.Match(input.text), "for input text %q", input.text)
				}
			}
		})
	}
}

func TestMatcher(t *testing.T) {
	explain := func(expr string) string {
		plan, err := Explain(expr)
		if err != nil {
			return "err: " + err.Error()
		}
		return plan
	}
	is := internal.Assert(t)
	{
		is.Eq("''", explain(""))
		is.Eq("//", explain("//"))
		for _, expr := range []string{"", "//"} {
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
	}
	{
		expr := "NOT //"
		is.Eq("NOT[//]", explain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.False(m.Match("a"))
	}
	{
		expr := "/^$/"
		is.Eq("/^$/", explain(expr))
		m := MustCompile(expr)
		is.True(m.Match(""))
		is.False(m.Match("a"))
	}
	{
		expr := "NOT /^$/"
		is.Eq("NOT[/^$/]", explain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.True(m.Match("a"))
	}
	{
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
	}
	{
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
	}
	{
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
	}
	{
		expr := "/^\\/home/ AND NOT /^\\/home\\/tmp/"
		is.Eq("AND[/^/home/,NOT[/^/home/tmp/]]", explain(expr))
		m := MustCompile(expr)
		is.True(m.Match("/home/joe/bin"))
		is.False(m.Match("/home/tmp/bin"))
		is.True(m.Match("/home/joe/tmp"))
		is.False(m.Match("/opt/home/joe/tmp"))
	}
	{
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
	}
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
