package settings

import (
	"hypr-dock/internal/pkg/conf"
	"hypr-dock/internal/pkg/flags"
	"hypr-dock/internal/pkg/pinned"
	"hypr-dock/internal/pkg/utils"

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

func Init(flags flags.Flags, log hclog.Logger) (*Settings, error) {
	var err error

	// get local app dir
	localDir := filepath.Join(os.Getenv("HOME"), LOCAL, APP_NAME)

	// read pinned file
	pinnedPath := filepath.Join(localDir, "pinned")
	pinnedApps, err := pinned.Open(pinnedPath)
	if err != nil {
		log.Error("Failed to create/write pinned list", "file", pinnedPath, "error", err)
	}

	// main configs dir
	configDir, isCreate, err := GetConfigDir(flags.DevMode)
	log.Debug("Config dir init", "path", configDir, "created", isCreate, "error", err)

	// main config file
	configPath := filepath.Join(configDir, APP_NAME+".conf")
	if flags.Config != "~/.config/hypr-dock" {
		configPath = expand(flags.Config)
	}

	// themes dir
	themesDir := filepath.Join(configDir, "themes")

	// read main config and current theme config
	config, err := conf.New(configPath, themesDir, log)
	if err != nil {
		log.Error("Confog faild", "error", err)
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

func GetConfigDir(dev bool) (string, bool, error) {
	var target, source string

	if dev {
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		configs := filepath.Join(filepath.Dir(exeDir), "configs")

		target = filepath.Join(configs, "dev")
		source = filepath.Join(configs, "default")
	} else {
		target = filepath.Join(os.Getenv("HOME"), ".config", APP_NAME)
		source = filepath.Join("/etc", APP_NAME)
	}

	created := utils.DirExists(target)

	if created {
		return target, created, nil
	}

	err := utils.CopyDir(target, source)
	if err != nil {
		return source, created, err
	}

	return target, created, nil
}

func expand(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
