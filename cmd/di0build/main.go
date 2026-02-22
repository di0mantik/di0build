package main

import (
	"os"

	"di0build/internal/config"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "~/.dotfiles/di0build.yaml"
	}

	cfg := config.MustLoad(cfgPath)
}
