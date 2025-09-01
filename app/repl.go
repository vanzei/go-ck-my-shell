package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type config struct {
	commandArgs []string
}

func cleanInput(text string) []string {
	output := strings.ToLower(text)
	words := strings.Fields(output)
	return words
}

func startRepl(cfg *config) {
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, "$ ") // <-- Print prompt before reading input
		reader.Scan()

		words := cleanInput(reader.Text())
		if len(words) == 0 {
			continue
		}
		commandName := words[0]

		args := words[1:]

		command, exists := getCommands()[commandName]
		if exists {
			cfg.commandArgs = args
			err := command.callback(cfg)
			if err != nil {
				fmt.Println(err)
			}
			continue

		} else {

			fmt.Println(commandName[:] + ": command not found")
			continue
		}
	}
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "exit is a shell builtin",
			callback:    commandExit,
		},
		"echo": {
			name:        "echo",
			description: "echo is a shell builtin",
			callback:    commandEcho,
		},
		"type": {
			name: "type",
			description: "type is a shell builtin",
			callback: commandType,
		},
	}
}
