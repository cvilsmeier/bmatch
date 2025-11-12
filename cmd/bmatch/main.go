// Cmd bmatch is a string matching tool with boolean expressions.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cvilsmeier/bmatch"
)

func usage() {
	fmt.Println("bmatch - a string matcher with grouping and boolean expressions")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("    bmatch [flags] expr [file]...")
	fmt.Println("")
	fmt.Println("    Bmatch reads the given files and prints matching lines.")
	fmt.Println("    If no files are given, it reads stdin.")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("")
	fmt.Println("    -explain")
	fmt.Println("            Print expression tree and exit.")
	fmt.Println("            Useful for hunting down shell escaping issues.")
	fmt.Println("    -lower")
	fmt.Println("            Convert all input lines to lowercase before matching.")
	fmt.Println("            Useful for ignoring case.")
	fmt.Println("    -help")
	fmt.Println("            Print this help page and exit")
	fmt.Println("")
	fmt.Println("https://github.com/cvilsmeier/bmatch")
}

func main() {
	var explain bool
	var lower bool
	flag.Usage = usage
	flag.BoolVar(&explain, "explain", explain, "")
	flag.BoolVar(&lower, "lower", lower, "")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Usage: bmatch [flags] expr [file]...")
		fmt.Println("Try 'bmatch -help' for more information.")
		os.Exit(1)
		return
	}
	expr := flag.Arg(0)
	if explain {
		plan, err := bmatch.Explain(expr)
		if err != nil {
			fmt.Printf("bmatch: %s\n", err)
			os.Exit(1)
			return
		}
		fmt.Printf("%s\n", plan)
		return
	}
	matcher, err := bmatch.Compile(expr)
	if err != nil {
		fmt.Printf("bmatch: %s\n", err)
		os.Exit(1)
		return
	}
	if flag.NArg() == 1 {
		matchReader(os.Stdin, matcher, lower)
	}
	for i := range flag.NArg() - 1 {
		filename := flag.Arg(i + 1)
		matchFile(filename, matcher, lower)
	}
}

func matchFile(filename string, matcher bmatch.Matcher, lower bool) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	defer f.Close()
	matchReader(f, matcher, lower)
}

func matchReader(r io.Reader, matcher bmatch.Matcher, lower bool) {
	sca := bufio.NewScanner(r)
	for sca.Scan() {
		line := sca.Text()
		if matchLine(line, matcher, lower) {
			fmt.Println(line)
		}
	}
	err := sca.Err()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}
}

func matchLine(line string, matcher bmatch.Matcher, lower bool) bool {
	if lower {
		line = strings.ToLower(line)
	}
	return matcher.Match(line)
}
