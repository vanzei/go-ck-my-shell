package main

import (
	"strings"
	"unicode"
)

// Returns commandName, commandArgs
func parseInputWithQuotes(input string) (string, []string) {
	var tokens []string
	var current strings.Builder
	inSingle, inDouble, escaped := false, false, false

	for _, c := range input {
		switch {

		case escaped:
			current.WriteRune(c)
			escaped = false
		case c == '\\' && !inDouble && !inSingle:
			escaped = true

		case c == '\'' && !inDouble:
			inSingle = !inSingle

		case c == '"' && !inSingle:
			inDouble = !inDouble

		case unicode.IsSpace(c) && !inSingle && !inDouble:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}

		default:
			current.WriteRune(c)

		}
	}

	// Add the last token if present
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
