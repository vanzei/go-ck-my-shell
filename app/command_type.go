package main

import (
	"fmt"
)

func commandType(cfg *config) error {
	if len(cfg.commandArgs) == 0 {
		return nil
	}
	target := cfg.commandArgs[0]
	// First, check if it's a builtin
	if cmd, ok := getCommands()[target]; ok {
		fmt.Println(cmd.description)
		return nil
	}
	// Otherwise, search in $PATH
	return handlerSearchFile(cfg, target)
}
