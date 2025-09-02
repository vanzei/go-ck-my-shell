package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// findExecutablesInDir searches for executable files within a given directory.
func findExecutablesInDir(dirPath string) ([]string, error) {
	var executables []string

	// Use filepath.WalkDir for efficient directory traversal.
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Return errors encountered during traversal.
		}

		// Skip directories themselves.
		if d.IsDir() {
			return nil
		}

		// Get file info to check permissions.
		info, err := d.Info()
		if err != nil {
			return err
		}

		// Check if the file is executable.
		// On Unix-like systems, this means checking the execute bit.
		// On Windows, executables typically have a .exe extension.
		// This example focuses on checking the execute bit, common for general executables.
		if info.Mode().IsRegular() && (info.Mode().Perm()&0o111 != 0) { // Check for execute permissions for owner, group, or others
			executables = append(executables, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", dirPath, err)
	}

	return executables, nil
}

func splitPath() []string {
	// Get the value of the PATH environment variable
	pathValue := os.Getenv("PATH")

	// Check if the PATH variable is set
	if pathValue == "" {
		fmt.Println("The PATH environment variable is not set.")
		return []string{}
	}

	// Use os.PathListSeparator for correct delimiter
	pathDirs := strings.Split(pathValue, string(os.PathListSeparator))

	return pathDirs
}

func handlerSearchFile(cfg *config, target string) error {
	paths := splitPath()
	for _, dir := range paths {
		fullPath := filepath.Join(dir, target)
		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() && info.Mode().Perm()&0o111 != 0 {
			fmt.Printf("%s is %s\n", target, fullPath)
			return nil
		}
	}
	//fmt.Printf("%s: not found\n", target)
	return fmt.Errorf("%s: not found", target)
}
