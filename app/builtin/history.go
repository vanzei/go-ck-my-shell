package builtin

import (
	"fmt"
	"io"
)

func commandHistory(cfg *Config, w io.Writer) error {
	for i, entry := range cfg.History {
		fmt.Fprintf(w, "%d  %s\n", i+1, entry)
	}
	return nil
}
