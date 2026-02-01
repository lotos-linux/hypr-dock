package settings

import (
	"hypr-dock/internal/pkg/conf"
	"hypr-dock/internal/pkg/flags"
	"hypr-dock/internal/pkg/pinned"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
)

const APP_NAME = "hypr-dock"
const LOCAL = ".local/share"

type Settings struct {
	*conf.Config
	LocalDir   string
	ConfigDir  string
	ConfigPath string
	PinnedPath string
	ThemesDir  string
	ThemeStyle string
	PinnedApps []string
}

func Init(flags flags.Flags, logger hclog.Logger) (*Settings, error) {
	var err error

	// get local app dir
	localDir := filepath.Join(GetHome(), LOCAL, APP_NAME)

	// read pinned file
	pinnedPath := filepath.Join(localDir, "pinned")
	pinnedApps, err := pinned.Open(pinnedPath)
	if err != nil {
		log.Fatal(err)
	}

	// main configs dir
	configDir := getConfigDir(flags.DevMode)

	// main config file
	configPath := filepath.Join(configDir, APP_NAME+".conf")
	if flags.Config != "~/.config/hypr-dock" {
		configPath = expand(flags.Config)
	}

	// themes dir
	themesDir := filepath.Join(configDir, "themes")

	// read main config and current theme config
	config, err := conf.New(configPath, themesDir, logger)
	if err != nil {
		log.Fatal(err)
	}

	// theme style file
	themeStyle := filepath.Join(config.ThemeDir, "style.css")

	return &Settings{
		Config:     config,
		LocalDir:   localDir,
		ConfigDir:  configDir,
		ConfigPath: configPath,
		PinnedPath: pinnedPath,
		ThemesDir:  themesDir,
		ThemeStyle: themeStyle,
		PinnedApps: pinnedApps,
	}, nil
}

func getConfigDir(dev bool) string {
	if dev {
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		return filepath.Join(filepath.Dir(exeDir), "configs")
	}

	home := GetHome()
	return filepath.Join(home, ".config", APP_NAME)
}

func expand(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func GetHome() string {
	home, _ := os.UserHomeDir()
	return home
}
