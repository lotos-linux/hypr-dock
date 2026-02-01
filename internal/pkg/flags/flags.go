package flags

import (
	"flag"
	"os"
)

type Flags struct {
	DevMode bool
	Config  string
	Theme   string
}

func Get() Flags {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	dev := flag.Bool("dev", false, "enable developer mode")
	config := flag.String("config", "~/.config/hypr-dock", "config file")
	theme := flag.String("theme", "", "theme dir")
	flag.Parse()

	return Flags{
		DevMode: *dev,
		Config:  *config,
		Theme:   *theme,
	}
}
