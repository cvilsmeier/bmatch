# bmatch

Bmatch is a Go (golang) string matching package and command line tool
that supports grouping and boolean expressions:

    /DEBUG/ OR ( /TRACE/ AND NOT /TRACE.*sql/ )

The expression syntax is:

    expr        := literalExpr | groupExpr | boolExpr
    groupExpr   := "(" expr ")"
    boolExpr    := not | and | or
    notExpr     := "NOT" expr
    andExpr     := expr "AND" expr
    orExpr      := expr "OR" expr
    literal     := "/" regex "/"
    regex       := ? a valid golang regex ?

The operator precedence is the same as in C (the programming language):

    - group <-- highest precedence
    - NOT
    - AND
    - OR    <-- lowest precedence


## Usage as library

    $ go get github.com/cvilsmeier/bmatch@latest


```go
package main

import (
	"bmatch"
	"fmt"
)

func main() {
	matcher := bmatch.MustCompile("/debug/ OR ( /trace/ AND NOT /trace.*sql/ )")
	fmt.Println(matcher.Match("debug some"))     // true
	fmt.Println(matcher.Match("trace some"))     // true
	fmt.Println(matcher.Match("trace some sql")) // false
}
```


## Install as command line tool

    $ go install github.com/cvilsmeier/bmatch/cmd@latest
    
    $ bmatch --help
    
    bmatch - a string matcher with grouping and boolean expressions

    Usage:

        bmatch [flags] expr [file]...

        Bmatch reads the given files and prints matching lines.
        If no files are given, it reads stdin.

    Flags:

        -explain
                Print expression tree and exit.
                Useful for hunting down shell escaping issues.
        -help
                Print this help page and exit


## License

~~~
MIT License

Copyright (c) 2025 Christoph Vilsmeier

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
~~~
