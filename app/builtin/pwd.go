package builtin

import (
	"fmt"
	"os"
)

func CommandPwd(cfg *Config) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}
