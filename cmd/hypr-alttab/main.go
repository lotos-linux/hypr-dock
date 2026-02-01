package main

import (
	"flag"
	"hypr-dock/internal/settings"
	"hypr-dock/internal/switcher"
	"os"
	"path/filepath"
)

func main() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	dev := flag.Bool("dev", false, "enable developer mode")

	configDir := settings.GetConfigDir(*dev)
	configPath := filepath.Join(configDir, "hypr-alttab.conf")

	switcher.Run(configPath)
}
