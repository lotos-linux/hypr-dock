package flags

import (
	"flag"
	"os"
)

type Flags struct {
	DevMode  bool
	Config   string
	Theme    string
	LogLevel string
}

func Get() Flags {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	dev := flag.Bool("dev", false, "enable developer mode")
	config := flag.String("config", "~/.config/hypr-dock", "config file")
	theme := flag.String("theme", "", "theme dir")
	logLevel := flag.String("log-level", "info", "log level")
	flag.Parse()

	return Flags{
		DevMode:  *dev,
		Config:   *config,
		Theme:    *theme,
		LogLevel: *logLevel,
	}
}
