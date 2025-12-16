package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	builtinPkg "github.com/codecrafters-io/shell-starter-go/app/builtin"
	"github.com/codecrafters-io/shell-starter-go/app/core"
	execPkg "github.com/codecrafters-io/shell-starter-go/app/exec"
	parserPkg "github.com/codecrafters-io/shell-starter-go/app/parser"
	"golang.org/x/term"
)

var errInterrupt = errors.New("interrupt")

type lineEditor struct {
	trie       *core.Trie
	lastPrefix string
	tabCount   int
	prompt     string
	history    []string
	historyPos int
}

func newLineEditor(commands []string, prompt string) *lineEditor {
	trie := core.NewTrie()
	for _, cmd := range commands {
		trie.Insert(cmd)
	}
	return &lineEditor{trie: trie, prompt: prompt}
}

func (e *lineEditor) setHistory(hist []string) {
	e.history = hist
	e.historyPos = len(hist)
}

func (e *lineEditor) resetTab(prefix string) {
	if prefix != e.lastPrefix {
		e.tabCount = 0
	}
	e.lastPrefix = prefix
}

func (e *lineEditor) handleCompletion(buf []rune) []rune {
	prefix := string(buf)
	matches := e.trie.FindWordsWithPrefix(prefix)

	e.resetTab(prefix)

	if len(matches) == 0 {
		fmt.Fprint(os.Stdout, "\a")
		e.tabCount = 0
		return buf
	}

	lcp := longestCommonPrefix(matches)
	if lcp != prefix {
		add := []rune(lcp[len(prefix):])
		fmt.Print(string(add))
		e.tabCount = 0
		e.lastPrefix = lcp
		return append(buf, add...)
	}

	if len(matches) == 1 && matches[0] == prefix {
		fmt.Print(" ")
		e.tabCount = 0
		return append(buf, ' ')
	}

	e.tabCount++
	if e.tabCount == 1 {
		fmt.Fprint(os.Stdout, "\a")
		return buf
	}
	if e.tabCount == 2 {
		fmt.Print("\r\n" + joinWithDoubleSpace(matches) + "\r\n" + e.prompt + string(buf))
		e.tabCount = 0
	}
	return buf
}

func (e *lineEditor) rewriteLine(buf []rune, prevLen int) {
	fmt.Print("\r")
	fmt.Print(e.prompt)
	fmt.Print(string(buf))
	if prevLen > len(buf) {
		fmt.Print(strings.Repeat(" ", prevLen-len(buf)))
		fmt.Print("\r")
		fmt.Print(e.prompt)
		fmt.Print(string(buf))
	}
}

func (e *lineEditor) readLine() (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(e.prompt)

	var buf []rune
	e.historyPos = len(e.history)
	prevLen := 0
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}
		switch r {
		case '\r', '\n':
			fmt.Print("\r\n")
			e.tabCount = 0
			e.lastPrefix = ""
			return string(buf), nil
		case '\t':
			buf = e.handleCompletion(buf)
		case 127, '\b':
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b")
			}
			e.tabCount = 0
			e.lastPrefix = string(buf)
		case 3: // Ctrl+C
			fmt.Print("^C\r\n")
			e.tabCount = 0
			e.lastPrefix = ""
			return "", errInterrupt
		case 4: // Ctrl+D
			if len(buf) == 0 {
				fmt.Print("\r\n")
				return "", io.EOF
			}
		case 0x1b: // basic escape sequence swallowing (arrows, etc.)
			seq, _ := reader.Peek(2)
			if len(seq) == 2 && seq[0] == '[' {
				reader.Discard(2)
				switch seq[1] {
				case 'A': // Up arrow
					if len(e.history) == 0 {
						continue
					}
					if e.historyPos > 0 {
						e.historyPos--
					}
					prevLen = len(buf)
					buf = []rune(e.history[e.historyPos])
					e.rewriteLine(buf, prevLen)
					e.tabCount = 0
					e.lastPrefix = string(buf)
				case 'B': // Down arrow
					if len(e.history) == 0 {
						continue
					}
					if e.historyPos < len(e.history) {
						e.historyPos++
					}
					prevLen = len(buf)
					if e.historyPos == len(e.history) {
						buf = []rune{}
					} else {
						buf = []rune(e.history[e.historyPos])
					}
					e.rewriteLine(buf, prevLen)
					e.tabCount = 0
					e.lastPrefix = string(buf)
				default:
				}
				continue
			}
		default:
			buf = append(buf, r)
			fmt.Print(string(r))
			e.tabCount = 0
			e.lastPrefix = string(buf)
		}
	}
}

// Helper to join matches with two spaces
func joinWithDoubleSpace(matches []string) string {
	return strings.Join(matches, "  ")
}

// Add this helper function:
func longestCommonPrefix(strs []string) string {
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
	return prefix
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

	editor := newLineEditor(commands, "$ ")

	defer func() {
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

	for {
		line, err := editor.readLine()
		if err == errInterrupt {
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
