package main

import (
	"fmt"

	builtinPkg "github.com/codecrafters-io/shell-starter-go/app/builtin"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	// Wait for user input
	var _ = fmt.Fprint

	cfg := &builtinPkg.Config{}
	startRepl(cfg)
}
