package bmatch

import (
	"testing"
)

func TestMatcher(t *testing.T) {
	t.Run("matchEverything", func(t *testing.T) {
		is := assert(t)
		for _, expr := range []string{"", "//"} {
			is.eq("//", MustExplain(expr))
			m := MustCompile(expr)
			is.true(m.Match(""))
			is.true(m.Match("a"))
			is.true(m.Match("ababab"))
			is.true(m.Match("aa"))
			is.true(m.Match("bb"))
			is.true(m.Match("aabb"))
			is.true(m.Match("bb  aa"))
			is.true(m.Match("bb  aabb  aa"))
		}
	})
	t.Run("matchNothing", func(t *testing.T) {
		is := assert(t)
		expr := "NOT //"
		is.eq("NOT[//]", MustExplain(expr))
		m := MustCompile(expr)
		is.false(m.Match(""))
		is.false(m.Match("a"))
	})
	t.Run("matchOnlyEmptyString", func(t *testing.T) {
		is := assert(t)
		expr := "/^$/"
		is.eq("/^$/", MustExplain(expr))
		m := MustCompile(expr)
		is.true(m.Match(""))
		is.false(m.Match("a"))
	})
	t.Run("matchOnlyNonEmptyString", func(t *testing.T) {
		is := assert(t)
		expr := "NOT /^$/"
		is.eq("NOT[/^$/]", MustExplain(expr))
		m := MustCompile(expr)
		is.false(m.Match(""))
		is.true(m.Match("a"))
	})
	t.Run("andOperator", func(t *testing.T) {
		is := assert(t)
		expr := "/aa/ AND /bb/"
		is.eq("AND[/aa/,/bb/]", MustExplain(expr))
		m := MustCompile(expr)
		is.false(m.Match(""))
		is.false(m.Match("a"))
		is.false(m.Match("ababab"))
		is.false(m.Match("aa"))
		is.false(m.Match("bb"))
		is.true(m.Match("aabb"))
		is.true(m.Match("bb  aa"))
		is.true(m.Match("bb  aabb  aa"))
	})
	t.Run("andOr", func(t *testing.T) {
		is := assert(t)
		expr := "/foo/ AND /bar/ OR /baz/"
		is.eq("OR[AND[/foo/,/bar/],/baz/]", MustExplain(expr))
		m := MustCompile(expr)
		is.false(m.Match(""))
		is.false(m.Match("foo"))
		is.false(m.Match("bar"))
		is.true(m.Match("baz"))
		is.true(m.Match("foobar"))
		is.true(m.Match("foobaz"))
		is.true(m.Match("barbaz"))
		is.true(m.Match("foobarbaz"))
		is.true(m.Match("barfoobaz"))
		is.true(m.Match("bazbarfoo"))
	})
	t.Run("andOrGroup", func(t *testing.T) {
		is := assert(t)
		expr := "/foo/ AND (/bar/ OR /baz/)"
		is.eq("AND[/foo/,OR[/bar/,/baz/]]", MustExplain(expr))
		m := MustCompile(expr)
		is.false(m.Match(""))
		is.false(m.Match("foo"))
		is.false(m.Match("bar"))
		is.false(m.Match("baz"))
		is.true(m.Match("foobar"))
		is.true(m.Match("foobaz"))
		is.false(m.Match("barbaz"))
		is.true(m.Match("foobarbaz"))
		is.true(m.Match("barfoobaz"))
		is.true(m.Match("bazbarfoo"))
	})
	t.Run("readme", func(t *testing.T) {
		is := assert(t)
		expr := "/DEBUG/ OR ( /TRACE/ AND NOT /TRACE.*sql/ )"
		is.eq("OR[/DEBUG/,AND[/TRACE/,NOT[/TRACE.*sql/]]]", MustExplain(expr))
		m := MustCompile(expr)
		is.true(m.Match("10:00 DEBUG will poll now"))
		is.true(m.Match("10:01 DEBUG polling error: no route to host"))
		is.true(m.Match("10:02 TRACE(poll) connecting 10.0.0.21"))
		is.false(m.Match("10:03 TRACE(sql) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.false(m.Match("10:04 TRACE(sql) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.true(m.Match("10:05 TRACE(http) POST /contactForm from 174.161.32.109"))
		is.false(m.Match("10:06 TRACE(sql) UPDATE sessions SET lastAccess=? WHERE id=?"))
		is.true(m.Match("11:00 DEBUG will poll now"))
	})
}
