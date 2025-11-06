// Package bmatch is a string matching library with boolean expressions.
package bmatch

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cvilsmeier/bmatch/internal"
)

// A Matcher matches strings.
type Matcher interface {
	Match(str string) bool
}

// MustCompile is like Compile but panics on error.
func MustCompile(expr string) Matcher {
	m, err := Compile(expr)
	if err != nil {
		panic(fmt.Sprintf("bmatch: Compile(%q): %s", expr, err))
	}
	return m
}

// Compile parses a bmatch expression and returns, if successful,
// a [Matcher] object that can be used to match strings.
func Compile(expr string) (Matcher, error) {
	node, err := compileNode(expr)
	if err != nil {
		return nil, err
	}
	return buildMatcher(0, node)
}

// Explain parses a bmatch expression and returns, if successful,
// a string representation of its syntax tree.
func Explain(expr string) (string, error) {
	node, err := compileNode(expr)
	if err != nil {
		return "", err
	}
	return explainNode(0, node), nil
}

func compileNode(expr string) (internal.Node, error) {
	lex, err := internal.NewStringLexer(expr)
	if err != nil {
		return internal.Node{}, err
	}
	return internal.Parse(lex)
}

func explainNode(level int, node internal.Node) string {
	var str string
	switch node.Typ {
	case internal.StringNode:
		str = "'" + node.Text + "'"
	case internal.RegexNode:
		str = "/" + node.Text + "/"
	case internal.NotNode:
		str = "NOT"
	case internal.AndNode:
		str = "AND"
	case internal.OrNode:
		str = "OR"
	default:
		panic("bad node typ")
	}
	if len(node.Subnodes) > 0 {
		str += "["
		for i, child := range node.Subnodes {
			if i > 0 {
				str += ","
			}
			str += explainNode(level+1, child)
		}
		str += "]"
	}
	return str
}

const maxLevels = 20

func buildMatcher(level int, node internal.Node) (Matcher, error) {
	if level > maxLevels {
		return nil, fmt.Errorf("too deep nesting level %d", level)
	}
	var submatchers []Matcher
	for _, subnode := range node.Subnodes {
		submatcher, err := buildMatcher(level+1, subnode)
		if err != nil {
			return nil, err
		}
		submatchers = append(submatchers, submatcher)
	}
	switch node.Typ {
	case internal.StringNode:
		return &stringMatcher{node.Text}, nil
	case internal.RegexNode:
		rex, err := regexp.Compile(node.Text)
		if err != nil {
			return nil, err
		}
		return &regexMatcher{rex}, nil
	case internal.NotNode:
		return &notMatcher{submatchers}, nil
	case internal.AndNode:
		return &andMatcher{submatchers}, nil
	case internal.OrNode:
		return &orMatcher{submatchers}, nil
	default:
		panic("bad node type")
	}
}

// A stringMatcher matches if the input contains a given string.
type stringMatcher struct {
	str string
}

func (m *stringMatcher) Match(str string) bool {
	return strings.Contains(str, m.str)
}

// A regexMatcher matches if the input matches a given regular expression.
type regexMatcher struct {
	rex *regexp.Regexp
}

func (m *regexMatcher) Match(str string) bool {
	return m.rex.MatchString(str)
}

// A notMatcher matches if no child matcher matches.
type notMatcher struct {
	matchers []Matcher
}

func (m *notMatcher) Match(str string) bool {
	for _, child := range m.matchers {
		if child.Match(str) {
			return false
		}
	}
	return true
}

// An andMatcher matches if every child matcher matches.
type andMatcher struct {
	matchers []Matcher
}

func (m *andMatcher) Match(str string) bool {
	for _, child := range m.matchers {
		if !child.Match(str) {
			return false
		}
	}
	return true
}

// An orMatcher matches if at least one child matcher matches.
type orMatcher struct {
	matchers []Matcher
}

func (m *orMatcher) Match(str string) bool {
	for _, child := range m.matchers {
		if child.Match(str) {
			return true
		}
	}
	return false
}
