package main

import (
	"strings"
	"unicode"
)

func parseInputWithQuotes(input string) (string, []string) {
	var tokens []string
	var current strings.Builder
	inSingle, inDouble := false, false

	for i := 0; i < len(input); i++ {
		c := rune(input[i])

		switch {
		case c == '\\' && !inSingle && !inDouble:
			// Outside quotes: escape next character
			if i+1 < len(input) {
				i++
				current.WriteByte(input[i])
			}
		case inDouble && c == '\\':
			// Inside double quotes: only escape ", \, $, `
			if i+1 < len(input) && strings.ContainsRune("\"\\$`", rune(input[i+1])) {
				i++
				current.WriteByte(input[i])
			} else if i+1 < len(input) {
				// For other characters, keep the backslash
				current.WriteByte(byte(c))
				i++
				current.WriteByte(input[i])
			}
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
