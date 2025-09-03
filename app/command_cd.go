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
	err := os.Chdir(path)
	if err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", path)
	}
	return nil
}
