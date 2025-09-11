package main

import (
	"bufio"
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
	histFile := os.Getenv("HISTFILE")
	// Load history from HISTFILE on startup
	if histFile != "" {
		file, err := os.Open(histFile)
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
	}

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
		HistoryFile:     histFile,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "readline error:", err)
		return
	}
	defer func() {
		rl.Close()
		histFile := os.Getenv("HISTFILE")
		if histFile != "" {
			file, err := os.OpenFile(histFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				for _, entry := range cfg.History {
					fmt.Fprintln(file, entry)
				}
				file.Close()
			}
		}
	}()

	cfg.RL = rl

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			break
		}

		cfg.History = append(cfg.History, line)

		c := line
		segments := parserPkg.SplitCommandsRespectingQuotes(c)
		numCmds := len(segments)
		if numCmds == 0 {
			continue
		}
		// Set up pipes
		pipes := make([][2]*os.File, numCmds-1)
		for i := 0; i < numCmds-1; i++ {
			r, w, err := os.Pipe()
			if err != nil {
				fmt.Fprintln(os.Stderr, "pipe error:", err)
				continue
			}
			pipes[i][0] = r
			pipes[i][1] = w
		}

		cmds := make([]*exec.Cmd, numCmds)
		outFiles := make([]*os.File, numCmds)
		errFiles := make([]*os.File, numCmds)

		for i, segment := range segments {
			commandName, args := parserPkg.ParseInputWithQuotes(segment)
			var outFile, errorFile *os.File
			var outputWriter io.Writer = os.Stdout
			var errorWriter io.Writer = os.Stderr

			// Redirection handling (same as before)
			j := 0
			for j < len(args) {
				switch args[j] {
				case ">", "1>":
					if j+1 < len(args) {
						f, err := os.Create(args[j+1])
						if err == nil {
							outFile = f
							outputWriter = outFile
						}
						args = append(args[:j], args[j+2:]...)
						continue
					}
				case ">>", "1>>":
					if j+1 < len(args) {
						f, err := os.OpenFile(args[j+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
						if err == nil {
							outFile = f
							outputWriter = outFile
						}
						args = append(args[:j], args[j+2:]...)
						continue
					}
				case "2>":
					if j+1 < len(args) {
						f, err := os.Create(args[j+1])
						if err == nil {
							errorFile = f
							errorWriter = errorFile
						}
						args = append(args[:j], args[j+2:]...)
						continue
					}
				case "2>>":
					if j+1 < len(args) {
						f, err := os.OpenFile(args[j+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
						if err == nil {
							errorFile = f
							errorWriter = errorFile
						}
						args = append(args[:j], args[j+2:]...)
						continue
					}
				}
				j++
			}
			outFiles[i] = outFile
			errFiles[i] = errorFile

			command, existsInternal := builtinPkg.GetCommands()[commandName]
			if existsInternal {
				// Builtin: only support as first or last in pipeline for now
				cfg.CommandArgs = args
				if i == 0 {
					if numCmds == 1 {
						err := command.CallbackWithWriter(cfg, outputWriter)
						if err != nil {
							fmt.Fprintln(errorWriter, err)
						}
					} else {
						// If builtin is first in pipeline, output to pipe
						err := command.CallbackWithWriter(cfg, pipes[0][1])
						if err != nil {
							fmt.Fprintln(errorWriter, err)
						}
						pipes[0][1].Close()
					}
				} else if i == numCmds-1 {
					err := command.CallbackWithWriter(cfg, outputWriter)
					if err != nil {
						fmt.Fprintln(errorWriter, err)
					}
				}
				if outFile != nil {
					outFile.Close()
				}
				if errorFile != nil {
					errorFile.Close()
				}
				continue
			}

			_, err := execPkg.HandlerSearchFile(cfg, commandName)
			if err != nil {
				fmt.Fprintln(errorWriter, err)
			}

			cmd := exec.Command(commandName, args...)
			// Set up stdin/stdout for pipeline
			if i == 0 {
				cmd.Stdin = os.Stdin
			} else {
				cmd.Stdin = pipes[i-1][0]
			}
			if i == numCmds-1 {
				cmd.Stdout = outputWriter
			} else {
				cmd.Stdout = pipes[i][1]
			}
			cmd.Stderr = errorWriter
			cmds[i] = cmd
		}

		// Start all commands
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			err := cmd.Start()
			if err != nil {
				continue
				// fmt.Fprintln(os.Stderr, "error starting command:", err)
			}
		}
		// Close write ends in parent
		for i := 0; i < numCmds-1; i++ {
			pipes[i][1].Close()
		}
		// Wait for all commands
		for i, cmd := range cmds {
			if cmd == nil {
				continue
			}
			err := cmd.Wait()
			if err != nil {
				// fmt.Fprintln(os.Stderr, "error waiting for command:", err)
				continue
			}
			if outFiles[i] != nil {
				outFiles[i].Close()
			}
			if errFiles[i] != nil {
				errFiles[i].Close()
			}
		}
		// Close read ends
		for i := 0; i < numCmds-1; i++ {
			pipes[i][0].Close()
		}
	}
}
