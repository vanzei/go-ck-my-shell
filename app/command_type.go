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
	fullPath, err := handlerSearchFile(cfg, target)
	if err != nil {
		return err
	}

	fmt.Printf("%s is %s\n", target, fullPath)
	return nil
}
