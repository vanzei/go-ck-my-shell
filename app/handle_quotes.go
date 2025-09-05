package main

import (
	"strings"
)

// Returns commandName, commandArgs
func parseInputWithQuotes(input string) (string, []string) {
	var tokens []string
	var current strings.Builder
	inSingle, inDouble := false, false

	for i := 0; i < len(input); i++ {
		c := input[i]
		switch c {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
		case ' ':
			if !inSingle && !inDouble {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				continue
			}
		}
		current.WriteByte(c)
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	// Strip quotes from tokens
	for i, t := range tokens {
		if len(t) > 1 && ((t[0] == '"' && t[len(t)-1] == '"') || (t[0] == '\'' && t[len(t)-1] == '\'')) {
			tokens[i] = t[1 : len(t)-1]
		}
	}

	if len(tokens) == 0 {
		return "", nil
	}
	return tokens[0], tokens[1:]
}
