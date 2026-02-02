package internal

import (
	"strings"
	"testing"
)

func TestStringLexer(t *testing.T) {
	type testcase struct {
		name  string
		input string
		want  string
	}
	for _, tt := range []testcase{
		// empty input
		{"empty_1", "", ""},
		{"empty_2", "          ", ""},
		// parentheses
		{"paren_01", "(", "("},
		{"paren_02", " ( ", "("},
		{"paren_03", ")", ")"},
		{"paren_04", "()", "(, )"},
		{"paren_05", "( )", "(, )"},
		{"paren_06", "(())", "(, (, ), )"},
		{"paren_07", "( ( ) )", "(, (, ), )"},
		{"paren_08", "(()())", "(, (, ), (, ), )"},
		{"paren_09", "( ( ) ( ) )", "(, (, ), (, ), )"},
		{"paren_10", " (   ()()  )  ", "(, (, ), (, ), )"},
		// string literals
		{"string_01", "a", "                  'a'"},
		{"string_05", "a b c", "              'a', 'b', 'c'"},
		{"string_06", "a\\ b\\ c", "          'a b c'"},
		{"string_10", "a b) c", "             'a', 'b', ), 'c'"},
		{"string_15", "a b\\) c", "           'a', 'b)', 'c'"},
		{"string_20", "a \\ b c", "           'a', ' b', 'c'"},
		{"string_25", "a b\\ b c", "          'a', 'b b', 'c'"},
		{"string_30", "(\\/a\\/)\\(b\\)", "   (, '/a/', ), '(b)'"},
		// regex literals
		{"regex_01", "//", "                        r[]"},
		{"regex_02", "////", "                      r[], r[]"},
		{"regex_03", "// //", "                     r[], r[]"},
		{"regex_11", "/a/", "                       r[a]"},
		{"regex_11", "/a b c/", "                   r[a b c]"},
		{"regex_12", "/a\\/", "                     err: unclosed regex in \"a/\""},
		{"regex_13", "/a\\//", "                    r[a/]"},
		{"regex_21", "/aa/", "                      r[aa]"},
		{"regex_22", "/a a/", "                     r[a a]"},
		{"regex_23", "  /a a/  ", "                 r[a a]"},
		{"regex_24", "/a//b/", "                    r[a], r[b]"},
		{"regex_25", "/a/ /b/", "                   r[a], r[b]"},
		{"regex_31", " /a/   /b/  ", "              r[a], r[b]"},
		{"regex_32", "/ / /b/", "                   r[ ], r[b]"},
		{"regex_33", "/aa/ /b/", "                  r[aa], r[b]"},
		{"regex_34", "(/a/)(/b/)", "                (, r[a], ), (, r[b], )"},
		{"regex_35", "  (  /a/)(   /b/   ) ", "     (, r[a], ), (, r[b], )"},
		{"regex_36", " ( /\\/a\\//)( /(b)/ ) ", "   (, r[/a/], ), (, r[(b)], )"},
		// NOT
		{"not_1", "NOT", "           NOT"},
		{"not_2", "NOT NOT", "       NOT, NOT"},
		{"not_3", "NOT //", "        NOT, r[]"},
		{"not_4", "NOT /aa/", "      NOT, r[aa]"},
		{"not_5", "NOT aa", "        NOT, 'aa'"},
		// AND
		{"and_01", "AND", "          AND"},
		{"and_02", "AND AND", "      AND, AND"},
		{"and_03", "AND //", "       AND, r[]"},
		{"and_11", "AND /a/", "      AND, r[a]"},
		{"and_12", "AND aaa", "      AND, 'aaa'"},
		{"and_21", "AND NOT /a/", "  AND, NOT, r[a]"},
		{"and_22", "AND NOT aaa", "  AND, NOT, 'aaa'"},
		// OR
		{"or_1", "OR", "             OR"},
		{"or_2", "OR OR", "          OR, OR"},
		// combinations
		{"combi_01", "/a/", "                                    r[a]"},
		{"combi_01", "a", "                                      'a'"},
		{"combi_02", "/a/ AND /b/", "                            r[a], AND, r[b]"},
		{"combi_03", "a AND b", "                                'a', AND, 'b'"},
		{"combi_11", "/a/AND/b/", "                              r[a], AND, r[b]"},
		{"combi_12", "/a/AND b", "                               r[a], AND, 'b'"},
		{"combi_13", "/a/ AND /b/ OR /c/", "                     r[a], AND, r[b], OR, r[c]"},
		{"combi_14", "a AND /b/ OR c", "                         'a', AND, r[b], OR, 'c'"},
		{"combi_15", "(/a/OR/b/)AND(/c/)", "                     (, r[a], OR, r[b], ), AND, (, r[c], )"},
		{"combi_21", "  (  /a/OR  /b/ )    AND (/c/)    ", "     (, r[a], OR, r[b], ), AND, (, r[c], )"},
		{"combi_22", "/a/ OR /b/ AND  /c/ AND NOT /d/ ", "       r[a], OR, r[b], AND, r[c], AND, NOT, r[d]"},
		{"combi_23", "(/a/ OR /b/) AND ( /c/ AND NOT /d/ )", "   (, r[a], OR, r[b], ), AND, (, r[c], AND, NOT, r[d], )"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := Assert(t)
			lex, err := NewStringLexer(tt.input)
			want := strings.TrimSpace(tt.want)
			var have string
			if err != nil {
				have = "err: " + err.Error()
			} else {
				have = collectAndDumpForTest(lex)
			}
			if have != want {
				is.Eq(want, have)
			}
		})
	}
}

func collectAndDumpForTest(lex Lexer) string {
	var toks []string
	for range 100 {
		t, err := lex.NextToken()
		if err != nil {
			panic("unexpected err in nextToken()")
		}
		switch t.Typ {
		case OpenToken:
			toks = append(toks, "(")
		case CloseToken:
			toks = append(toks, ")")
		case NotToken:
			toks = append(toks, "NOT")
		case AndToken:
			toks = append(toks, "AND")
		case OrToken:
			toks = append(toks, "OR")
		case StringToken:
			toks = append(toks, "'"+t.Text+"'")
		case RegexToken:
			toks = append(toks, "r["+t.Text+"]")
		case EOFToken:
			return strings.Join(toks, ", ")
		default:
			panic("bad token typ")
		}
	}
	return "too many tokens"
}
