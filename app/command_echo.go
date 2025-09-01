package main

import "fmt"

func commandEcho(cfg *config) error {
	for _, arg := range cfg.commandArgs {
		fmt.Print(arg + " ")
	}
	fmt.Println()

	return nil
}
