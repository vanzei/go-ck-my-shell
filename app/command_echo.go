package main

import (
	"fmt"
	"io"
)

func commandEcho(cfg *config, w io.Writer) error {
	for i, arg := range cfg.commandArgs {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, arg)
	}
	fmt.Fprint(w, "\n")
	return nil
}
