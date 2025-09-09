package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type config struct {
	commandArgs []string
}
type shellCompleter struct {
	commands   []string
	lastPrefix string
	tabCount   int
}

func (c *shellCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	word := string(line[:pos])
	var matches [][]rune
	for _, cmd := range c.commands {
		if strings.HasPrefix(cmd, word) {
			matches = append(matches, []rune(cmd))
		}
	}

	// Reset tabCount if prefix changed
	if word != c.lastPrefix {
		c.tabCount = 0
	}
	c.lastPrefix = word

	if len(matches) == 1 {
		// Only one match: autocomplete the missing part and add a space
		completion := append(matches[0][len(word):], ' ')
		c.tabCount = 0
		return [][]rune{completion}, pos
	}
	if len(matches) == 0 {
		fmt.Fprint(os.Stdout, "\a")
		c.tabCount = 0
		return nil, pos
	}
	// Multiple matches
	c.tabCount++
	if c.tabCount == 1 {
		// First tab: emit bell
		fmt.Fprint(os.Stdout, "\a")
		return nil, pos
	}
	if c.tabCount == 2 {
		// Second tab: print matches separated by 2 spaces, then prompt and buffer
		fmt.Fprintln(os.Stdout, "\n"+joinWithDoubleSpace(matches))
		fmt.Fprint(os.Stdout, "$ "+word)
		c.tabCount = 0 // reset for next time
		return nil, pos
	}
	return nil, pos
}

// Helper to join matches with two spaces
func joinWithDoubleSpace(matches [][]rune) string {
	var strs []string
	for _, m := range matches {
		strs = append(strs, string(m))
	}
	return strings.Join(strs, "  ")
}

func startRepl(cfg *config) {
	// Use a map to deduplicate command names
	unique := make(map[string]struct{})

	// Add built-in commands
	for cmd := range getCommands() {
		unique[cmd] = struct{}{}
	}

	// Add external commands
	executables := allExecutables()
	for _, exec := range executables {
		unique[exec] = struct{}{}
	}

	// Build the deduplicated slice
	var commands []string
	for cmd := range unique {
		commands = append(commands, cmd)
	}

	// Sort commands alphabetically
	sort.Strings(commands)

	// Set up completer
	completer := &shellCompleter{commands: commands}
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
					f, err := os.OpenFile(args[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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
					f, err := os.OpenFile(args[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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
