package builtin

import (
	"io"
)

type CliCommand struct {
	Name               string
	Description        string
	CallbackWithWriter func(*Config, io.Writer) error
}

func GetCommands() map[string]CliCommand {
	return map[string]CliCommand{
		"exit": {
			Name:        "exit",
			Description: "exit is a shell builtin",
			CallbackWithWriter: func(cfg *Config, w io.Writer) error {
				return CommandExit(cfg)
			},
		},
		"echo": {
			Name:               "echo",
			Description:        "echo is a shell builtin",
			CallbackWithWriter: CommandEcho,
		},
		"type": {
			Name:        "type",
			Description: "type is a shell builtin",
			CallbackWithWriter: func(cfg *Config, w io.Writer) error {
				return CommandType(cfg)
			},
		},
		"pwd": {
			Name:        "pwd",
			Description: "pwd is a shell builtin",
			CallbackWithWriter: func(cfg *Config, w io.Writer) error {
				return CommandPwd(cfg)
			},
		},
		"cd": {
			Name:        "cd",
			Description: "cd is a shell builtin",
			CallbackWithWriter: func(cfg *Config, w io.Writer) error {
				return CommandCd(cfg)
			},
		},
		"history": {
			Name:        "history",
			Description: "history is a shell builtin",
			CallbackWithWriter: func(cfg *Config, w io.Writer) error {
				return commandHistory(cfg, w)
			},
		},
	}
}
