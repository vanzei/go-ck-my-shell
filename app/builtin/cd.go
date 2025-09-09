package builtin

import (
	"fmt"
	"os"
)

func CommandCd(cfg *Config) error {
	if len(cfg.CommandArgs) < 1 {
		return nil
	}
	err := ChangeDirectory(cfg.CommandArgs[0])
	if err != nil {
		return err
	}
	return nil
}

func ChangeDirectory(path string) error {
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
