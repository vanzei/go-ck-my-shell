package builtin

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func commandHistory(cfg *Config, w io.Writer) error {
	// Handle history -r <file>
	if len(cfg.CommandArgs) >= 2 && cfg.CommandArgs[0] == "-r" {
		file, err := os.Open(cfg.CommandArgs[1])
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					cfg.History = append(cfg.History, line)
				}
			}
			file.Close()
		}
		return nil // Do not print history after reading
	}
	// Handle history -w <file>
	if len(cfg.CommandArgs) >= 2 && cfg.CommandArgs[0] == "-w" {
		file, err := os.Create(cfg.CommandArgs[1])
		if err == nil {
			for _, entry := range cfg.History {
				fmt.Fprintln(file, entry)
			}
			file.Close()
		}
		return nil // Do not print history after writing
	}
	// Handle history -a <file>
	if len(cfg.CommandArgs) >= 2 && cfg.CommandArgs[0] == "-a" {
		file, err := os.OpenFile(cfg.CommandArgs[1], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err == nil {
			// Track last appended index in cfg (add LastHistoryAppend int to Config)
			start := 0
			if cfg.LastHistoryAppend > 0 {
				start = cfg.LastHistoryAppend
			}
			for i := start; i < len(cfg.History); i++ {
				fmt.Fprintln(file, cfg.History[i])
			}
			file.Close()
			cfg.LastHistoryAppend = len(cfg.History)
		}
		return nil // Do not print history after appending
	}
	// Print in-memory history if available
	if len(cfg.History) > 0 {
		for i, entry := range cfg.History {
			fmt.Fprintf(w, "%d  %s\n", i+1, entry)
		}
		return nil
	}
	// Otherwise, print HISTFILE if set
	histFile := os.Getenv("HISTFILE")
	if histFile != "" {
		file, err := os.Open(histFile)
		if err == nil {
			scanner := bufio.NewScanner(file)
			i := 1
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Fprintf(w, "%d  %s\n", i, line)
				i++
			}
			file.Close()
		}
	}
	return nil
}
