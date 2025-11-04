package bmatch_test

import (
	"fmt"
	"log"

	"github.com/cvilsmeier/bmatch"
)

func ExampleCompile() {
	matcher, err := bmatch.Compile("/foo/ AND NOT /bar/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(matcher.Match("foo"))     // true
	fmt.Println(matcher.Match("bar"))     // false
	fmt.Println(matcher.Match("foobar"))  // false
	fmt.Println(matcher.Match("barfoo"))  // false
	fmt.Println(matcher.Match("foother")) // true
	// Output:
	// true
	// false
	// false
	// false
	// true
}

func ExampleExplain() {
	plan, err := bmatch.Explain("/foo/ OR /bar/ AND NOT /bill/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(plan)
	// Output: OR[/foo/,AND[/bar/,NOT[/bill/]]]
}
