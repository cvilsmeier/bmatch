# bmatch

[![GoDoc Reference](https://pkg.go.dev/badge/github.com/cvilsmeier/bmatch)](http://godoc.org/github.com/cvilsmeier/bmatch)
[![Build Status](https://github.com/cvilsmeier/bmatch/actions/workflows/build.yml/badge.svg)](https://github.com/cvilsmeier/bmatch/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Bmatch is a Go (golang) string matching package and command line tool that
supports boolean expressions:

    /debug/ OR ( /trace/ AND NOT /trace.*sql/ )

The expression syntax is (whitespace ignored for simplicity):

    <expr>          ::=  <literal> | <operator>
    <operator>      ::=  <groupExpr> | <notExpr> | <andExpr> | <orExpr>
    <groupExpr>     ::=  "(" <expr> ")"
    <notExpr>       ::=  "NOT" <expr>
    <andExpr>       ::=  <expr> "AND" <expr>
    <orExpr>        ::=  <expr> "OR" expr
    <literal>       ::=  <stringLiteral> || <regexLiteral>
    <stringLiteral> ::=  ? any character string not enclosed in "/" characters ?
    <regexLiteral>  ::=  "/" <regex> "/"
    <regex>         ::=  ? any valid golang regex ?

The operator precedence is the same as in C (the programming language):

    - NOT   <-- highest precedence
    - AND
    - OR    <-- lowest precedence

The precedence can be changed by using parentheses.



## Usage

~~~
go get github.com/cvilsmeier/bmatch
~~~


```go
func main() {
	matcher := bmatch.MustCompile("/debug/ OR ( /trace/ AND NOT /trace.*sql/ )")
	fmt.Println(matcher.Match("debug some"))     // true
	fmt.Println(matcher.Match("trace some"))     // true
	fmt.Println(matcher.Match("trace some sql")) // false
}
```


## Command line tool

~~~
go install github.com/cvilsmeier/bmatch/cmd/bmatch@latest
~~~

~~~
$ bmatch - a string matcher with boolean expressions

Usage:

    bmatch [flags] expr [file]...

    Bmatch reads the given files and prints matching lines.
    If no files are given, it reads stdin.

Flags:

    -explain
            Print expression tree and exit.
            Useful for hunting down shell escaping issues.
    -lower
            Convert all input lines to lowercase before matching.
            Useful for ignoring case.
    -help
            Print this help page and exit
~~~
