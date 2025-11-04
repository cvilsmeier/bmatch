package bmatch

import (
	"fmt"
	"slices"
)

// Parse input tokens and build an abstract syntax tree.
// It's a L-R parser with a one-token lookahead.
func parse(lex lexer) (node, error) {
	stack := &stack{nil}
	lookahead, err := lex.nextToken()
	if err != nil {
		return node{}, err
	}
	if lookahead.isEOF() {
		// fast path for empty input: match empty regex (in other words: match exerything)
		return node{ntLiteral, "", nil}, nil
	}
	const maxTokens = 1000 // prevent endless loop
	for range maxTokens {
		// process next token
		token := lookahead
		lookahead, err = lex.nextToken()
		if err != nil {
			return node{}, err
		}
		// shift (push token onto stack)
		stack.push(token)
		// reduce (build nodes from stack)
		stack.reduce(lookahead)
		// did it terminate?
		if lookahead.typ == ttEOF {
			if stack.len() == 1 {
				item := stack.first()
				if item.isNode() {
					return item.node, nil
				}
			}
			return node{}, fmt.Errorf("syntax error")
		}
	}
	return node{}, fmt.Errorf("too many tokens")
}

type node struct {
	typ      nodeTyp
	text     string
	subnodes []node
}

func (n node) isZero() bool { return int(n.typ) == 0 }

type nodeTyp int

const (
	_ nodeTyp = iota
	ntLiteral
	ntNot
	ntAnd
	ntOr
)

func (t nodeTyp) String() string {
	switch t {
	case ntLiteral:
		return "ntLiteral"
	case ntNot:
		return "ntNot"
	case ntAnd:
		return "ntAnd"
	case ntOr:
		return "ntOr"
	}
	return fmt.Sprintf("NodeTyp(%d)", int(t))
}

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

func (s *stack) push(token token) {
	s.items = append(s.items, stackitem{token: token})
}

// reduce reduces the stack by creating nodes according to the
// following reduction rules:
//
//	literal          --> node
//	"NOT" node       --> node
//	node "AND" node  --> node
//	node "OR" node   --> (lookahead "OR", ")", EOF)  -->  node
//	"(" node ")"     --> node
func (s *stack) reduce(lookahead token) error {
	const maxRounds = 100
	for range maxRounds {
		if s.reduceLiteralToken() {
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
		if s.reduceOpenCloseToken() {
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
		if item.isTokenOf(ttLiteral) {
			newNode := node{typ: ntLiteral, text: item.token.text}
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
		if i1.isTokenOf(ttNot) && i2.isNode() {
			newNode := node{typ: ntNot, text: i1.token.text, subnodes: []node{i2.node}}
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
		if i1.isNode() && i2.isTokenOf(ttAnd) && i3.isNode() {
			newNode := node{typ: ntAnd, text: i2.token.text, subnodes: []node{i1.node, i3.node}}
			s.replaceItems(nitems-3, nitems, newNode)
			return true
		}
	}
	return false
}

func (s *stack) reduceOrToken(lookahead token) bool {
	switch lookahead.typ {
	case ttClose, ttOr, ttEOF:
		nitems := len(s.items)
		if nitems >= 3 {
			i1 := s.items[nitems-3] // node
			i2 := s.items[nitems-2] // OR
			i3 := s.items[nitems-1] // node
			if i1.isNode() && i2.isTokenOf(ttOr) && i3.isNode() {
				newNode := node{typ: ntOr, text: i2.token.text, subnodes: []node{i1.node, i3.node}}
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
		if i1.isTokenOf(ttOpen) && i2.isNode() && i3.isTokenOf(ttClose) {
			s.replaceItems(nitems-3, nitems, i2.node)
			return true
		}
	}
	return false
}

func (s *stack) replaceItems(from, to int, node node) {
	if from != to {
		s.items = slices.Delete(s.items, from+1, to)
	}
	s.items[from] = stackitem{node: node}
}

// A stack item holds either a token or a node.
type stackitem struct {
	token token
	node  node
}

func (si stackitem) isToken() bool               { return !si.token.isZero() }
func (si stackitem) isTokenOf(typ tokenTyp) bool { return si.isToken() && si.token.typ == typ }
func (si stackitem) isNode() bool                { return !si.node.isZero() }
