package exec

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FindExecutablesInDir(dirPath string, w io.Writer) []string {
	var executables []string
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && (info.Mode().Perm()&0o111 != 0) {
			executables = append(executables, path)
		}
		return nil
	})
	if err != nil {
		return []string{"error walking directory "}
	}
	return executables
}

func SplitPath() []string {
	pathValue := os.Getenv("PATH")
	if pathValue == "" {
		fmt.Println("The PATH environment variable is not set.")
		return []string{}
	}
	pathDirs := strings.Split(pathValue, string(os.PathListSeparator))
	return pathDirs
}

func HandlerSearchFile(cfg interface{}, target string) (string, error) {
	paths := SplitPath()
	for _, dir := range paths {
		fullPath := filepath.Join(dir, target)
		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() && info.Mode().Perm()&0o111 != 0 {
			return fullPath, nil
		}
	}
	return "", fmt.Errorf("%s: not found", target)
}
