package bmatch

import (
	"fmt"
)

// A lexer yields tokens, one after another.
// The last token is an EOF token.
type lexer interface {
	nextToken() (token, error)
}

type token struct {
	typ  tokenTyp
	text string
}

func (t token) isZero() bool { return int(t.typ) == 0 }
func (t token) isEOF() bool  { return t.typ == ttEOF }

type tokenTyp int

const (
	_ tokenTyp = iota
	ttOpen
	ttClose
	ttNot
	ttAnd
	ttOr
	ttLiteral
	ttEOF
)

func (t tokenTyp) String() string {
	switch t {
	case ttOpen:
		return "ttOpen"
	case ttClose:
		return "ttClose"
	case ttNot:
		return "ttNot"
	case ttAnd:
		return "ttAnd"
	case ttOr:
		return "ttOr"
	case ttLiteral:
		return "ttLiteral"
	case ttEOF:
		return "ttEOF"
	}
	return fmt.Sprintf("TokenTyp(%d)", int(t))
}

// A stringLexer tokenizes an input string.
type stringLexer struct {
	tokens []token
}

func newStringLexer(input string) (*stringLexer, error) {
	var tokens []token
	var rbuf rstack
	var inEscape bool
	var inLiteral bool
	var inOperator bool
	for _, r := range input {
		if inEscape {
			rbuf.push(r)
			switch r {
			case '/':
				inEscape = false
			default:
				return nil, fmt.Errorf("invalid escape sequence in %q", rbuf.pop())
			}
		} else if inLiteral {
			switch r {
			case '/':
				inLiteral = false
				tokens = append(tokens, token{ttLiteral, rbuf.pop()})
			case '\\':
				inEscape = true
			default:
				rbuf.push(r)
			}
		} else {
			if inOperator {
				switch r {
				case ' ', '(', ')', '/':
					inOperator = false
					text := rbuf.pop()
					switch text {
					case "NOT":
						tokens = append(tokens, token{ttNot, ""})
					case "AND":
						tokens = append(tokens, token{ttAnd, ""})
					case "OR":
						tokens = append(tokens, token{ttOr, ""})
					default:
						return nil, fmt.Errorf("unknown operator %q", text)
					}
				}
			}
			switch r {
			case ' ':
				// space character are separators but carry no meaning
			case '(':
				tokens = append(tokens, token{ttOpen, ""})
			case ')':
				tokens = append(tokens, token{ttClose, ""})
			case '/':
				inLiteral = true
			default:
				rbuf.push(r)
				inOperator = true
			}
		}
	}
	if inLiteral {
		return nil, fmt.Errorf("unclosed literal")
	}
	if inOperator {
		text := rbuf.pop()
		switch text {
		case "NOT":
			tokens = append(tokens, token{ttNot, ""})
		case "AND":
			tokens = append(tokens, token{ttAnd, ""})
		case "OR":
			tokens = append(tokens, token{ttOr, ""})
		default:
			return nil, fmt.Errorf("unknown operator %q", text)
		}
	}
	return &stringLexer{tokens}, nil
}

func (l *stringLexer) nextToken() (token, error) {
	if len(l.tokens) == 0 {
		return token{ttEOF, ""}, nil
	}
	t := l.tokens[0]
	l.tokens = l.tokens[1:]
	return t, nil
}

// rstack is a stack of runes.
type rstack struct {
	b []rune
}

func (b *rstack) push(r rune) {
	b.b = append(b.b, r)
}

func (b *rstack) pop() string {
	text := string(b.b)
	b.b = nil
	return text
}
