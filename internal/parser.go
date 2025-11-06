package internal

import (
	"fmt"
	"slices"
)

// Parse input tokens and build an abstract syntax tree.
// It's a L-R parser with a one-token lookahead.
func Parse(lex Lexer) (Node, error) {
	stack := &stack{nil}
	lookahead, err := lex.NextToken()
	if err != nil {
		return Node{}, err
	}
	if lookahead.IsEOF() {
		// fast path for empty input: match empty regex (in other words: match exerything)
		return Node{LiteralNode, "", nil}, nil
	}
	const maxTokens = 1000 // prevent endless loop
	for range maxTokens {
		// process next token
		token := lookahead
		lookahead, err = lex.NextToken()
		if err != nil {
			return Node{}, err
		}
		// shift (push token onto stack)
		stack.push(token)
		// reduce (build nodes from stack)
		stack.reduce(lookahead)
		// did it terminate?
		if lookahead.Typ == TEOF {
			if stack.len() == 1 {
				item := stack.first()
				if item.isNode() {
					return item.node, nil
				}
			}
			return Node{}, fmt.Errorf("syntax error")
		}
	}
	return Node{}, fmt.Errorf("too many tokens")
}

// A Node is a node in the parse tree.
type Node struct {
	Typ      NodeTyp
	Text     string
	Subnodes []Node
}

func (n Node) isZero() bool { return int(n.Typ) == 0 }

type NodeTyp int

const (
	_ NodeTyp = iota
	LiteralNode
	NotNode
	AndNode
	OrNode
)

// A stack holds stack items, which can be tokens or nodes.
type stack struct {
	items []stackitem
}

func (s *stack) len() int {
	return len(s.items)
}

func (s *stack) first() stackitem {
	return s.items[0]
}

func (s *stack) push(token Token) {
	s.items = append(s.items, stackitem{token: token})
}

// reduce reduces the stack by creating nodes according to the
// following reduction rules:
//
//	literal          --> node
//	"(" node ")"     --> node
//	"NOT" node       --> node
//	node "AND" node  --> node
//	node "OR" node   --> (lookahead "OR", ")", EOF)  -->  node
func (s *stack) reduce(lookahead Token) error {
	const maxRounds = 100
	for range maxRounds {
		if s.reduceLiteralToken() {
			continue
		}
		if s.reduceOpenCloseToken() {
			continue
		}
		if s.reduceNotToken() {
			continue
		}
		if s.reduceAndToken() {
			continue
		}
		if s.reduceOrToken(lookahead) {
			continue
		}
		return nil
	}
	return fmt.Errorf("too many reduce rounds")
}

func (s *stack) reduceLiteralToken() bool {
	nitems := len(s.items)
	if nitems >= 1 {
		item := s.items[nitems-1]
		if item.isTokenOf(TLiteral) {
			newNode := Node{Typ: LiteralNode, Text: item.token.Text}
			s.replaceItems(nitems-1, nitems-1, newNode)
			return true
		}
	}
	return false
}

func (s *stack) reduceNotToken() bool {
	nitems := len(s.items)
	if nitems >= 2 {
		i1 := s.items[nitems-2] // NOT
		i2 := s.items[nitems-1] // node
		if i1.isTokenOf(TNot) && i2.isNode() {
			newNode := Node{Typ: NotNode, Text: i1.token.Text, Subnodes: []Node{i2.node}}
			s.replaceItems(nitems-2, nitems, newNode)
			return true
		}
	}
	return false
}

func (s *stack) reduceAndToken() bool {
	nitems := len(s.items)
	if nitems >= 3 {
		i1 := s.items[nitems-3] // node
		i2 := s.items[nitems-2] // AND
		i3 := s.items[nitems-1] // node
		if i1.isNode() && i2.isTokenOf(TAnd) && i3.isNode() {
			newNode := Node{Typ: AndNode, Text: i2.token.Text, Subnodes: []Node{i1.node, i3.node}}
			s.replaceItems(nitems-3, nitems, newNode)
			return true
		}
	}
	return false
}

func (s *stack) reduceOrToken(lookahead Token) bool {
	switch lookahead.Typ {
	case TClose, TOr, TEOF:
		nitems := len(s.items)
		if nitems >= 3 {
			i1 := s.items[nitems-3] // node
			i2 := s.items[nitems-2] // OR
			i3 := s.items[nitems-1] // node
			if i1.isNode() && i2.isTokenOf(TOr) && i3.isNode() {
				newNode := Node{Typ: OrNode, Text: i2.token.Text, Subnodes: []Node{i1.node, i3.node}}
				s.replaceItems(nitems-3, nitems, newNode)
				return true
			}
		}
	}
	return false
}

func (s *stack) reduceOpenCloseToken() bool {
	nitems := len(s.items)
	if nitems >= 3 {
		i1 := s.items[nitems-3] // (
		i2 := s.items[nitems-2] // node
		i3 := s.items[nitems-1] // )
		if i1.isTokenOf(TOpen) && i2.isNode() && i3.isTokenOf(TClose) {
			s.replaceItems(nitems-3, nitems, i2.node)
			return true
		}
	}
	return false
}

func (s *stack) replaceItems(from, to int, node Node) {
	if from != to {
		s.items = slices.Delete(s.items, from+1, to)
	}
	s.items[from] = stackitem{node: node}
}

// A stack item holds either a token or a node.
type stackitem struct {
	token Token
	node  Node
}

func (si stackitem) isToken() bool               { return !si.token.IsZero() }
func (si stackitem) isTokenOf(typ TokenTyp) bool { return si.isToken() && si.token.Typ == typ }
func (si stackitem) isNode() bool                { return !si.node.isZero() }
