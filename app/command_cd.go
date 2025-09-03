package main

import (
	"fmt"
	"os"
)

func commandCd(cfg *config) error {
	if len(cfg.commandArgs) < 1 {
		return nil
	}
	err := changeDirectory(cfg.commandArgs[0])
	if err != nil {
		return err
	}
	return nil
}

func changeDirectory(path string) error {
	if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cd: could not determine home directory")
		}
		path = homeDir
	}
	err := os.Chdir(path)
	if err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", path)
	}
	return nil
}
