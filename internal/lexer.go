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
func (t Token) IsEOF() bool  { return t.Typ == EOFToken }

type TokenTyp int

const (
	_ TokenTyp = iota
	OpenToken
	CloseToken
	NotToken
	AndToken
	OrToken
	StringToken
	RegexToken
	EOFToken
)

// A StringLexer tokenizes an input string.
type StringLexer struct {
	tokens []Token
}

func NewStringLexer(input string) (*StringLexer, error) {
	var stack rstack
	var tokens []Token
	consumeStack := func() {
		text := stack.pop()
		switch text {
		case "":
			// ignore
		case "NOT":
			tokens = append(tokens, Token{NotToken, ""})
		case "AND":
			tokens = append(tokens, Token{AndToken, ""})
		case "OR":
			tokens = append(tokens, Token{OrToken, ""})
		default:
			tokens = append(tokens, Token{StringToken, text})
		}
	}
	var inEscape bool
	var inRegex bool
	var inString bool
	for _, r := range input {
		if inEscape {
			switch r {
			case ' ', '(', ')', '/', '\\':
				stack.push(r)
				inEscape = false
			default:
				return nil, fmt.Errorf("invalid escape sequence in %q", stack.pop())
			}
		} else if inRegex {
			switch r {
			case '/':
				inRegex = false
				tokens = append(tokens, Token{RegexToken, stack.pop()})
			case '\\':
				inEscape = true
			default:
				stack.push(r)
			}
		} else if inString {
			switch r {
			case ' ':
				inString = false
				consumeStack()
			case '(':
				inString = false
				consumeStack()
				tokens = append(tokens, Token{OpenToken, ""})
			case ')':
				inString = false
				consumeStack()
				tokens = append(tokens, Token{CloseToken, ""})
			case '/':
				inString = false
				consumeStack()
				inRegex = true
			case '\\':
				inEscape = true
			default:
				stack.push(r)
			}
		} else {
			switch r {
			case ' ':
				// separator
			case '(':
				tokens = append(tokens, Token{OpenToken, ""})
			case ')':
				tokens = append(tokens, Token{CloseToken, ""})
			case '/':
				inRegex = true
			case '\\':
				inEscape = true
			default:
				stack.push(r)
				inString = true
			}
		}
	}
	if inEscape {
		return nil, fmt.Errorf("unclosed escape sequence in %q", stack.pop())
	}
	if inRegex {
		return nil, fmt.Errorf("unclosed regex in %q", stack.pop())
	}
	if inString {
		consumeStack()
	}
	return &StringLexer{tokens}, nil
}

func (l *StringLexer) NextToken() (Token, error) {
	if len(l.tokens) == 0 {
		return Token{EOFToken, ""}, nil
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
