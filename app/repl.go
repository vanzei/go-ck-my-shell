package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type config struct {
	commandArgs []string
}

func cleanInput(text string) []string {
	words := strings.Fields(text)
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

		command, existsInternal := getCommands()[commandName]
		if existsInternal {
			cfg.commandArgs = args
			err := command.callback(cfg)
			if err != nil {
				fmt.Println(err)
			}
			continue

		}

		_, err := handlerSearchFile(cfg, commandName)
		if err != nil {
			fmt.Println(err)
		}

		cmd := exec.Command(commandName, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
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
			name:        "type",
			description: "type is a shell builtin",
			callback:    commandType,
		},
		"pwd": {
			name:        "pwd",
			description: "pwd is a shell builtin",
			callback:    commandPwd,
		},
	}
}
