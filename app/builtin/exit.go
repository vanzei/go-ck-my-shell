package builtin

import "os"

func CommandExit(cfg *Config) error {
	os.Exit(0)
	return nil
}
