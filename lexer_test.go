package bmatch

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
		{"empty_1", "", "EOF"},
		{"empty_2", "          ", "EOF"},
		// parentheses
		{"paren_1", "(", "(,EOF"},
		{"paren_2", " ( ", "(,EOF"},
		{"paren_3", ")", "),EOF"},
		{"paren_4", "()", "(,),EOF"},
		{"paren_5", "(())", "(,(,),),EOF"},
		{"paren_6", "(()())", "(,(,),(,),),EOF"},
		{"paren_7", " (   ()()  )  ", "(,(,),(,),),EOF"},
		// literals
		{"lit_01", "//", "//,EOF"},
		{"lit_02", "////", "//,//,EOF"},
		{"lit_03", "// //", "//,//,EOF"},
		{"lit_04", "/a/", "/a/,EOF"},
		{"lit_05", "/aa/", "/aa/,EOF"},
		{"lit_06", "/a a/", "/a a/,EOF"},
		{"lit_07", "  /a a/  ", "/a a/,EOF"},
		{"lit_08", "/a//b/", "/a/,/b/,EOF"},
		{"lit_09", "/a/ /b/", "/a/,/b/,EOF"},
		{"lit_10", " /a/   /b/  ", "/a/,/b/,EOF"},
		{"lit_11", "/ / /b/", "/ /,/b/,EOF"},
		{"lit_12", "/aa/ /b/", "/aa/,/b/,EOF"},
		{"lit_13", "(/a/)(/b/)", "(,/a/,),(,/b/,),EOF"},
		{"lit_14", "   (   /a/)(       /b/    ) ", "(,/a/,),(,/b/,),EOF"},
		// NOT
		{"not_1", "NOT", "NOT,EOF"},
		{"not_2", "NOT NOT", "NOT,NOT,EOF"},
		{"not_3", "NOT //", "   NOT,//,EOF"},
		{"not_4", "NOT //", "   NOT,//,EOF"},
		// AND
		{"and_1", "AND", "AND,EOF"},
		{"and_2", "AND AND", "AND,AND,EOF"},
		{"and_3", "AND //", "AND,//,EOF"},
		{"and_4", "AND /a/", "AND,/a/,EOF"},
		{"and_5", "AND NOT /a/", "AND,NOT,/a/,EOF"},
		// OR
		{"or_1", "OR", "OR,EOF"},
		{"or_2", "OR OR", "OR,OR,EOF"},
		// combinations
		{"combi_01", "/a/ AND /b/", "   /a/,AND,/b/,EOF"},
		{"combi_02", "/a/AND/b/", "     /a/,AND,/b/,EOF"},
		{"combi_03", "/a/ AND /b/ OR /c/", "   /a/,AND,/b/,OR,/c/,EOF"},
		{"combi_04", "(/a/OR/b/)AND(/c/)", "                (,/a/,OR,/b/,),AND,(,/c/,),EOF"},
		{"combi_05", "  (  /a/OR  /b/ )    AND (/c/)    ", "(,/a/,OR,/b/,),AND,(,/c/,),EOF"},
		{"combi_06", "/a/ OR /b/ AND  /c/ AND NOT /d/ ", "       /a/,OR,/b/,AND,/c/,AND,NOT,/d/,EOF"},
		{"combi_07", "(/a/ OR /b/) AND ( /c/ AND NOT /d/ )", "       (,/a/,OR,/b/,),AND,(,/c/,AND,NOT,/d/,),EOF"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := Assert(t)

			lex, err := newStringLexer(tt.input)
			want := strings.TrimSpace(tt.want)
			var have string
			if err != nil {
				have = "err: " + err.Error()
			} else {
				have = collectAndDump(lex)
			}
			if have != want {
				is.Eq(want, have)
			}
		})
	}
}

func collectAndDump(lex lexer) string {
	var toks []string
	for range 100 {
		t, err := lex.nextToken()
		if err != nil {
			panic("unexpected err in nextToken()")
		}
		switch t.typ {
		case ttOpen:
			toks = append(toks, "(")
		case ttClose:
			toks = append(toks, ")")
		case ttNot:
			toks = append(toks, "NOT")
		case ttAnd:
			toks = append(toks, "AND")
		case ttOr:
			toks = append(toks, "OR")
		case ttLiteral:
			toks = append(toks, "/"+t.text+"/")
		case ttEOF:
			toks = append(toks, "EOF")
			return strings.Join(toks, ",")
		default:
			panic("bad token typ")
		}
	}
	toks = append(toks, "too many tokens")
	return strings.Join(toks, "")
}
