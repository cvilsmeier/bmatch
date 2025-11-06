package internal

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	type testcase struct {
		input string
		want  string
	}
	is := Assert(t)
	for _, tt := range []testcase{
		// the empty expression
		{"", ""},
		// text expressions
		{"a", "                    a"},
		{"a b", "                  err: syntax error"},
		// NOT
		{"NOT a", "                NOT[a]"},
		{"NOT NOT a", "            NOT[NOT[a]]"},
		{"NOT NOT NOT a", "        NOT[NOT[NOT[a]]]"},
		{"NOT", "                  err: syntax error"},
		{"a NOT", "                err: syntax error"},
		{"a NOT b", "              err: syntax error"},
		// AND
		{"AND", "                  err: syntax error"},
		{"AND AND", "              err: syntax error"},
		{"a AND", "                err: syntax error"},
		{"a AND b", "              AND[a,b]"},
		{"a AND AND b", "          err: syntax error"},
		{"a AND b AND", "          err: syntax error"},
		{"a AND b AND c", "        AND[AND[a,b],c]"},
		{"a AND b AND c AND d", "  AND[AND[AND[a,b],c],d]"},
		// NOT & AND
		{"NOT AND", "              err: syntax error"},
		{"AND NOT", "              err: syntax error"},
		{"a AND NOT", "            err: syntax error"},
		{"a AND NOT b", "          AND[a,NOT[b]]"},
		{"NOT a AND NOT b", "      AND[NOT[a],NOT[b]]"},
		{"NOT a AND NOT NOT b", "  AND[NOT[a],NOT[NOT[b]]]"},
		{"NOT a AND b", "          AND[NOT[a],b]"},
		// OR
		{"OR", "                   err: syntax error"},
		{"OR OR", "                err: syntax error"},
		{"a OR", "                 err: syntax error"},
		{"a OR b", "               OR[a,b]"},
		{"a OR OR b", "            err: syntax error"},
		{"a OR b OR", "            err: syntax error"},
		{"a OR b OR c", "          OR[OR[a,b],c]"},
		{"a OR b OR c OR d", "     OR[OR[OR[a,b],c],d]"},
		// NOT & AND & OR
		{"a AND b OR c", "         OR[AND[a,b],c]"},
		{"a AND b OR c AND", "     err: syntax error"},
		{"a AND b OR c AND d", "   OR[AND[a,b],AND[c,d]]"},
		{"a OR b AND c OR", "      err: syntax error"},
		{"a OR b AND c OR d", "    OR[OR[a,AND[b,c]],d]"},
		{"a OR NOT", "             err: syntax error"},
		{"a OR NOT b", "           OR[a,NOT[b]]"},
		{"a OR NOT b NOT e", "     err: syntax error"},
		{"a OR NOT b AND NOT e", " OR[a,AND[NOT[b],NOT[e]]]"},
		// Parentheses
		{"(", "                    err: syntax error"},
		{")", "                    err: syntax error"},
		{"( a )", "                a"},
		{"a )", "                  err: syntax error"},
		{"( a", "                  err: syntax error"},
		{"( ( a )", "              err: syntax error"},
		{"( a ) )", "              err: syntax error"},
		{"( a b )", "              err: syntax error"},
		{"( a ) ( b )", "          err: syntax error"},
		// Parentheses & NOT & AND & OR
		{"( a )", "                   a"},
		{"( NOT a )", "               NOT[a]"},
		{"( ( NOT a ) )", "           NOT[a]"},
		{"NOT ( a )", "               NOT[a]"},
		{"NOT ( ( a ) )", "           NOT[a]"},
		{"NOT ( a ) )", "             err: syntax error"},
		{"NOT ( ( a )", "             err: syntax error"},
		{"( a ) AND ( b )", "                    AND[a,b]"},
		{"( a ) AND ( NOT b )", "                AND[a,NOT[b]]"},
		{"( a AND b ) OR c", "                   OR[AND[a,b],c]"},
		{"a AND ( b OR c )", "                   AND[a,OR[b,c]]"},
		{"( a ) AND ( e OR f )", "               AND[a,OR[e,f]]"},
		{"( a OR b ) AND NOT ( c OR NOT d )", "                       AND[OR[a,b],NOT[OR[c,NOT[d]]]]"},
		{"( a OR ( b AND c ) ) AND ( NOT g OR NOT ( h AND i ) )", "   AND[OR[a,AND[b,c]],OR[NOT[g],NOT[AND[h,i]]]]"},
	} {
		t.Logf("testcase '%s'", tt.input)
		lex := newFakeLexer(tt.input)
		node, err := Parse(lex)
		want := strings.TrimSpace(tt.want)
		if strings.HasPrefix(want, "err: ") {
			if err == nil {
				is.Failf("testcase '%s': want err but was ok", tt.input)
			}
			want := want[5:]
			have := err.Error()
			if have != want {
				is.Failf("testcase '%s'\nwant error '%s'\nhave error '%s'", tt.input, want, have)
			}
		} else {
			if err != nil {
				is.Failf("testcase '%s': want ok but have error %s", tt.input, err)
			}
			have := dumpNode(0, node)
			if have != want {
				is.Failf("testcase '%s'\nwant %q\nhave %q", tt.input, want, have)
			}
		}
	}
}

func dumpNode(level int, node Node) string {
	if level > 100 {
		panic("dumpNode: too deep")
	}
	str := node.Text
	if len(node.Subnodes) > 0 {
		str += "["
		for i, child := range node.Subnodes {
			if i > 0 {
				str += ","
			}
			str += dumpNode(level+1, child)
		}
		str += "]"
	}
	return str
}

// A fakeLexer yields pre-defined tokens.
type fakeLexer struct {
	toks []string
}

func newFakeLexer(input string) *fakeLexer {
	var toks []string
	if input != "" {
		toks = strings.Split(input, " ")
	}
	return &fakeLexer{toks}
}

func (l *fakeLexer) NextToken() (Token, error) {
	if len(l.toks) == 0 {
		return Token{EOFToken, "EOF"}, nil
	}
	tok := l.toks[0]
	l.toks = l.toks[1:]
	switch tok {
	case "(":
		return Token{OpenToken, "("}, nil
	case ")":
		return Token{CloseToken, ")"}, nil
	case "NOT":
		return Token{NotToken, "NOT"}, nil
	case "AND":
		return Token{AndToken, "AND"}, nil
	case "OR":
		return Token{OrToken, "OR"}, nil
	}
	if strings.HasPrefix(tok, "/") && strings.HasSuffix(tok, "/") {
		tok = tok[1 : len(tok)-1]
		if len(tok) == 0 {
			panic("cannot have empty regex token")
		}
		return Token{RegexToken, tok}, nil
	}
	if len(tok) == 0 {
		panic("cannot have empty string token")
	}
	return Token{StringToken, tok}, nil
}
