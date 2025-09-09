package builtin

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/exec"
)

func CommandType(cfg *Config) error {
	if len(cfg.CommandArgs) == 0 {
		return nil
	}
	target := cfg.CommandArgs[0]
	// First, check if it's a builtin
	if cmd, ok := GetCommands()[target]; ok {
		fmt.Println(cmd.Description)
		return nil
	}
	// Otherwise, search in $PATH
	fullPath, err := exec.HandlerSearchFile(cfg, target)
	if err != nil {
		return err
	}
	fmt.Printf("%s is %s\n", target, fullPath)
	return nil
}
