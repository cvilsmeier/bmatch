package internal

import (
	"fmt"
)

// A Lexer yields tokens, one after another.
// The last token is an EOF token.
type Lexer interface {
	NextToken() (Token, error)
}

type Token struct {
	Typ  TokenTyp
	Text string
}

func (t Token) IsZero() bool { return int(t.Typ) == 0 }
func (t Token) IsEOF() bool  { return t.Typ == TEOF }

type TokenTyp int

const (
	_ TokenTyp = iota
	TOpen
	TClose
	TNot
	TAnd
	TOr
	TLiteral
	TEOF
)

// A StringLexer tokenizes an input string.
type StringLexer struct {
	tokens []Token
}

func NewStringLexer(input string) (*StringLexer, error) {
	var tokens []Token
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
				tokens = append(tokens, Token{TLiteral, rbuf.pop()})
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
						tokens = append(tokens, Token{TNot, ""})
					case "AND":
						tokens = append(tokens, Token{TAnd, ""})
					case "OR":
						tokens = append(tokens, Token{TOr, ""})
					default:
						return nil, fmt.Errorf("unknown operator %q", text)
					}
				}
			}
			switch r {
			case ' ':
				// space character are separators but carry no meaning
			case '(':
				tokens = append(tokens, Token{TOpen, ""})
			case ')':
				tokens = append(tokens, Token{TClose, ""})
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
			tokens = append(tokens, Token{TNot, ""})
		case "AND":
			tokens = append(tokens, Token{TAnd, ""})
		case "OR":
			tokens = append(tokens, Token{TOr, ""})
		default:
			return nil, fmt.Errorf("unknown operator %q", text)
		}
	}
	return &StringLexer{tokens}, nil
}

func (l *StringLexer) NextToken() (Token, error) {
	if len(l.tokens) == 0 {
		return Token{TEOF, ""}, nil
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
