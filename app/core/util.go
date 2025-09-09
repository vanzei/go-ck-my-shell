package core

import (
	"fmt"
	"os"
	"strings"
)

func SplitPath() []string {
	pathValue := os.Getenv("PATH")
	if pathValue == "" {
		fmt.Println("The PATH environment variable is not set.")
		return []string{}
	}
	pathDirs := strings.Split(pathValue, string(os.PathListSeparator))
	return pathDirs
}
