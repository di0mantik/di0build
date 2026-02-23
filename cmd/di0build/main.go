package main

import (
	"fmt"
	"os"

	"di0build/internal/config"
	"di0build/internal/model"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		homeDir, _ := os.UserHomeDir()
		cfgPath = fmt.Sprintf("%s/.dotfiles/di0build.yaml", homeDir)
	}

	cfg := config.MustLoad(cfgPath)
	model.RunInstaller(cfg)
}
