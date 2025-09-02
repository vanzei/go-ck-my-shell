package main

import (
	"fmt"
	"os"
)

func commandPwd(cfg *config) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}
