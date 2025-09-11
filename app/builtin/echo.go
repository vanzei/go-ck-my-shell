package builtin

import (
	"fmt"
	"io"
)

// type Config struct {
// 	CommandArgs       []string
// 	RL                *readline.Instance
// 	History           []string
// 	LastHistoryAppend int
// }

func CommandEcho(cfg *Config, w io.Writer) error {
	for i, arg := range cfg.CommandArgs {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, arg)
	}
	fmt.Fprint(w, "\n")
	return nil
}
