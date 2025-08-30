package main

import (
	"fmt"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	// Wait for user input
	var _ = fmt.Fprint
	for {

		cfg := &config{}
		startRepl(cfg)
	}
}
