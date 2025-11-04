package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

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
	fmt.Println("    -help")
	fmt.Println("            Print this help page and exit")
	fmt.Println("")
	fmt.Println("https://github.com/cvilsmeier/bmatch")
}

func main() {
	var explain bool
	flag.Usage = usage
	flag.BoolVar(&explain, "explain", explain, "")
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
		matchReader(os.Stdin, matcher)
	}
	for i := range flag.NArg() - 1 {
		filename := flag.Arg(i + 1)
		matchFile(filename, matcher)
	}
}

func matchFile(filename string, matcher bmatch.Matcher) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	defer f.Close()
	matchReader(f, matcher)
}

func matchReader(r io.Reader, matcher bmatch.Matcher) {
	sca := bufio.NewScanner(r)
	for sca.Scan() {
		line := sca.Text()
		if matcher.Match(line) {
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
