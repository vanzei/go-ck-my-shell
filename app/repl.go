package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/chzyer/readline"
)

type config struct {
	commandArgs []string
}

func startRepl(cfg *config) {
	// Build a list of command names for completion
	var commands []string
	for cmd := range getCommands() {
		commands = append(commands, cmd)
	}

	// Set up completer
	completer := readline.NewPrefixCompleter()
	for _, cmd := range commands {
		completer.Children = append(completer.Children, readline.PcItem(cmd))
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "readline error:", err)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			break
		}

		c := line
		commandName, args := parseInputWithQuotes(c)

		var outFile, errorFile *os.File
		var outputWriter io.Writer = os.Stdout
		var errorWriter io.Writer = os.Stderr

		i := 0
		for i < len(args) {
			switch args[i] {
			case ">", "1>":
				if i+1 < len(args) {
					f, err := os.Create(args[i+1])
					if err == nil {
						outFile = f
						outputWriter = outFile
					}
					args = append(args[:i], args[i+2:]...)
					continue
				}
			case ">>", "1>>":
				if i+1 < len(args) {
					f, err := os.OpenFile(args[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err == nil {
						outFile = f
						outputWriter = outFile
					}
					args = append(args[:i], args[i+2:]...)
					continue
				}
			case "2>":
				if i+1 < len(args) {
					f, err := os.Create(args[i+1])
					if err == nil {
						errorFile = f
						errorWriter = errorFile
					}
					args = append(args[:i], args[i+2:]...)
					continue
				}
			case "2>>":
				if i+1 < len(args) {
					f, err := os.OpenFile(args[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err == nil {
						errorFile = f
						errorWriter = errorFile
					}
					args = append(args[:i], args[i+2:]...)
					continue
				}
			}
			i++
		}

		command, existsInternal := getCommands()[commandName]
		if existsInternal {
			cfg.commandArgs = args
			err := command.callbackWithWriter(cfg, outputWriter)
			if err != nil {
				fmt.Fprintln(errorWriter, err)
			}
			if outFile != nil {
				outFile.Close()
			}
			if errorFile != nil {
				errorFile.Close()
			}
			continue
		}

		_, err = handlerSearchFile(cfg, commandName)
		if err != nil {
			fmt.Fprintln(errorWriter, err)
		}

		cmd := exec.Command(commandName, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = outputWriter
		cmd.Stderr = errorWriter

		err = cmd.Run()
		if outFile != nil {
			outFile.Close()
		}
		if errorFile != nil {
			errorFile.Close()
		}

		if err != nil {
			continue
		}
	}
}

type cliCommand struct {
	name        string
	description string
	// Add callbackWithWriter for built-ins
	callbackWithWriter func(*config, io.Writer) error
}

// Update getCommands to use callbackWithWriter for built-ins
func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "exit is a shell builtin",
			callbackWithWriter: func(cfg *config, w io.Writer) error {
				return commandExit(cfg)
			},
		},
		"echo": {
			name:               "echo",
			description:        "echo is a shell builtin",
			callbackWithWriter: commandEcho,
		},
		"type": {
			name:        "type",
			description: "type is a shell builtin",
			callbackWithWriter: func(cfg *config, w io.Writer) error {
				return commandType(cfg)
			},
		},
		"pwd": {
			name:        "pwd",
			description: "pwd is a shell builtin",
			callbackWithWriter: func(cfg *config, w io.Writer) error {
				return commandPwd(cfg)
			},
		},
		"cd": {
			name:        "cd",
			description: "cd is a shell builtin",
			callbackWithWriter: func(cfg *config, w io.Writer) error {
				return commandCd(cfg)
			},
		},
	}
}
