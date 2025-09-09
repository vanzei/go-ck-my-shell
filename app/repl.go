package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	builtinPkg "github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/core"
	execPkg "github.com/codecrafters-io/shell-starter-go/app/exec"
	parserPkg "github.com/codecrafters-io/shell-starter-go/app/parser"
)

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

	if len(matches) == 0 {
		fmt.Fprint(os.Stdout, "\a")
		c.tabCount = 0
		return nil, pos
	}

	// Find longest common prefix among matches
	lcp := word
	if len(matches) > 0 {
		lcp = longestCommonPrefix(matches)
	}

	if lcp != word {
		// If the longest common prefix is a full command, add space
		for _, m := range matches {
			if string(m) == lcp && len(matches) == 1 {
				completion := append([]rune(lcp[len(word):]), ' ')
				c.tabCount = 0
				return [][]rune{completion}, pos
			}
		}
		// Otherwise, autocomplete to longest common prefix (no space)
		completion := []rune(lcp[len(word):])
		c.tabCount = 0
		return [][]rune{completion}, pos
	}

	// If only one match and it's fully typed, add space
	if len(matches) == 1 && string(matches[0]) == word {
		completion := []rune{' '}
		c.tabCount = 0
		return [][]rune{completion}, pos
	}

	// Multiple matches, but no further completion possible
	c.tabCount++
	if c.tabCount == 1 {
		fmt.Fprint(os.Stdout, "\a")
		return nil, pos
	}
	if c.tabCount == 2 {
		fmt.Fprintln(os.Stdout, "\n"+joinWithDoubleSpace(matches))
		fmt.Fprint(os.Stdout, "$ "+word)
		c.tabCount = 0
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

// Add this helper function:
func longestCommonPrefix(strs [][]rune) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		i := 0
		for i < len(prefix) && i < len(s) && prefix[i] == s[i] {
			i++
		}
		prefix = prefix[:i]
		if len(prefix) == 0 {
			break
		}
	}
	return string(prefix)
}

func startRepl(cfg *builtinPkg.Config) {
	// Use a map to deduplicate command names
	unique := make(map[string]struct{})

	// Add built-in commands
	for cmd := range builtinPkg.GetCommands() {
		unique[cmd] = struct{}{}
	}

	// Add external commands
	paths := core.SplitPath()
	for _, dir := range paths {
		for _, execName := range execPkg.FindExecutablesInDir(dir, os.Stdout) {
			base := execName
			if strings.Contains(execName, "/") {
				base = execName[strings.LastIndex(execName, "/")+1:]
			}
			unique[base] = struct{}{}
		}
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
		commandName, args := parserPkg.ParseInputWithQuotes(c)

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

		command, existsInternal := builtinPkg.GetCommands()[commandName]
		if existsInternal {
			cfg.CommandArgs = args
			err := command.CallbackWithWriter(cfg, outputWriter)
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

		_, err = execPkg.HandlerSearchFile(cfg, commandName)
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
