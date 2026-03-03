package settings

import (
	"fmt"
	"hypr-dock/internal/pkg/conf"
	"hypr-dock/internal/pkg/flags"
	"hypr-dock/internal/pkg/pinned"
	"io"

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
	localDir := filepath.Join(GetHome(), LOCAL, APP_NAME)

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
	if !dev {
		systemDir := filepath.Join("/etc", APP_NAME)
		userDir := filepath.Join(GetHome(), ".config", APP_NAME)

		exist, err := HasDir(userDir, systemDir)
		if err != nil {
			return "", false, err
		}

		return userDir, exist, nil
	}

	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	configs := filepath.Join(filepath.Dir(exeDir), "configs")
	devDir := filepath.Join(configs, "dev")
	def := filepath.Join(configs, "default")

	exist, err := HasDir(devDir, def)
	if err != nil {
		return "", false, err
	}

	return devDir, exist, nil
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

func HasDir(dir string, sourceDir string) (bool, error) {
	if _, err := os.Stat(dir); err == nil {
		return true, nil
	}

	if _, err := os.Stat(sourceDir); err != nil {
		return false, fmt.Errorf("source directory does not exist: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, fmt.Errorf("failed to create directory: %w", err)
	}

	return false, filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
