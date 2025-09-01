package main

import (
	"errors"
	"fmt"
)

func commandType(cfg *config) error {
	if len(cfg.commandArgs) == 0 {
		return nil
	}
	target := cfg.commandArgs[0]
	for key, option := range getCommands() {
		if key == target {
			fmt.Println(option.description)
			return nil
		}
	}
	return errors.New(target + ": not found")
}
