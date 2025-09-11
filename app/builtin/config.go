package builtin

import (
	"github.com/chzyer/readline"
)

type Config struct {
	CommandArgs       []string
	RL                *readline.Instance
	History           []string
	LastHistoryAppend int
}
