package builtin

import (
	"fmt"
	"io"
	"strconv"
)

func commandHistory(cfg *Config, w io.Writer) error {
	history := cfg.History
	n := len(history) // default: show all

	// If an argument is provided, parse it as a number
	if len(cfg.CommandArgs) > 0 {
		if num, err := strconv.Atoi(cfg.CommandArgs[0]); err == nil && num < n {
			n = num
		}
	}

	start := len(history) - n
	if start < 0 {
		start = 0
	}
	for i := start; i < len(history); i++ {
		fmt.Fprintf(w, "%d  %s\n", i+1, history[i])
	}
	return nil
}
