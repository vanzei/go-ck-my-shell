package builtin

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

func commandHistory(cfg *Config, w io.Writer) error {
	// Handle history -r <file>
	if len(cfg.CommandArgs) >= 2 && cfg.CommandArgs[0] == "-r" {
		file, err := os.Open(cfg.CommandArgs[1])
		if err != nil {
			fmt.Fprintf(w, "history: cannot open file: %v\n", err)
			return nil
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				cfg.History = append(cfg.History, line)
			}
		}
		return nil
	}

	// Handle history -w <file>
	if len(cfg.CommandArgs) >= 2 && cfg.CommandArgs[0] == "-w" {
		file, err := os.Create(cfg.CommandArgs[1])
		if err != nil {
			fmt.Fprintf(w, "history: cannot write file: %v\n", err)
			return nil
		}
		defer file.Close()
		for _, entry := range cfg.History {
			fmt.Fprintln(file, entry)
		}
		return nil
	}

	// Optionally support history N
	n := len(cfg.History)
	if len(cfg.CommandArgs) == 1 {
		if num, err := strconv.Atoi(cfg.CommandArgs[0]); err == nil && num < n {
			n = num
		}
	}
	start := len(cfg.History) - n
	if start < 0 {
		start = 0
	}
	for i := start; i < len(cfg.History); i++ {
		fmt.Fprintf(w, "%d  %s\n", i+1, cfg.History[i])
	}
	return nil
}
