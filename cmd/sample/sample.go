package main

import (
	"fmt"

	"github.com/cvilsmeier/bmatch"
)

func main() {
	matcher := bmatch.MustCompile("/debug/ OR ( /trace/ AND NOT /trace.*sql/ )")
	fmt.Println(matcher.Match("debug some"))     // true
	fmt.Println(matcher.Match("trace some"))     // true
	fmt.Println(matcher.Match("trace some sql")) // false
}
