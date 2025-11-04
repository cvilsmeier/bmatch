package bmatch

import (
	"testing"
)

func TestMatcher(t *testing.T) {
	t.Run("matchEverything", func(t *testing.T) {
		is := Assert(t)
		for _, expr := range []string{"", "//"} {
			is.Eq("//", MustExplain(expr))
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
		is := Assert(t)
		expr := "NOT //"
		is.Eq("NOT[//]", MustExplain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.False(m.Match("a"))
	})
	t.Run("matchOnlyEmptyString", func(t *testing.T) {
		is := Assert(t)
		expr := "/^$/"
		is.Eq("/^$/", MustExplain(expr))
		m := MustCompile(expr)
		is.True(m.Match(""))
		is.False(m.Match("a"))
	})
	t.Run("matchOnlyNonEmptyString", func(t *testing.T) {
		is := Assert(t)
		expr := "NOT /^$/"
		is.Eq("NOT[/^$/]", MustExplain(expr))
		m := MustCompile(expr)
		is.False(m.Match(""))
		is.True(m.Match("a"))
	})
	t.Run("andOperator", func(t *testing.T) {
		is := Assert(t)
		expr := "/aa/ AND /bb/"
		is.Eq("AND[/aa/,/bb/]", MustExplain(expr))
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
		is := Assert(t)
		expr := "/foo/ AND /bar/ OR /baz/"
		is.Eq("OR[AND[/foo/,/bar/],/baz/]", MustExplain(expr))
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
		is := Assert(t)
		expr := "/foo/ AND (/bar/ OR /baz/)"
		is.Eq("AND[/foo/,OR[/bar/,/baz/]]", MustExplain(expr))
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
	t.Run("readme", func(t *testing.T) {
		is := Assert(t)
		expr := "/DEBUG/ OR ( /TRACE/ AND NOT /TRACE.*sql/ )"
		is.Eq("OR[/DEBUG/,AND[/TRACE/,NOT[/TRACE.*sql/]]]", MustExplain(expr))
		m := MustCompile(expr)
		is.True(m.Match("10:00 DEBUG will poll now"))
		is.True(m.Match("10:01 DEBUG polling error: no route to host"))
		is.True(m.Match("10:02 TRACE(poll) connecting 10.0.0.21"))
		is.False(m.Match("10:03 TRACE(sql) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.False(m.Match("10:04 TRACE(sql) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.True(m.Match("10:05 TRACE(http) POST /contactForm from 174.161.32.109"))
		is.False(m.Match("10:06 TRACE(sql) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.True(m.Match("11:00 DEBUG will poll now"))
	})
}
