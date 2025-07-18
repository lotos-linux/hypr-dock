package settings

import (
	"hypr-dock/internal/pkg/cfg"
	"hypr-dock/internal/pkg/flags"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/pkg/ipc"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	cfg.Config
	ConfigDir              string
	ConfigPath             string
	PinnedPath             string
	ThemesDir              string
	CurrentThemeDir        string
	CurrentThemeConfigPath string
	CurrentThemeStylePath  string
	PinnedApps             []string
}

func Init() (Settings, error) {
	const DefaultTheme = "lotos"

	var settings Settings

	flags := flags.Get(DefaultTheme)

	if flags.DevMode {
		settings.ConfigDir = setConfigDir("dev")
	} else {
		settings.ConfigDir = setConfigDir("normal")
	}

	settings.PinnedPath = filepath.Join(settings.ConfigDir, "pinned.json")
	settings.PinnedApps = cfg.ReadItemList(settings.PinnedPath)
	defaultConfigPath := filepath.Join(settings.ConfigDir, "config.jsonc")

	if flags.Config == "~/.config/hypr-dock" {
		settings.ConfigPath = defaultConfigPath
	} else {
		settings.ConfigPath = expandPath(flags.Config)
	}

	settings.Config = cfg.ReadConfig(settings.ConfigPath, settings.ThemesDir)

	settings.ThemesDir = filepath.Join(settings.ConfigDir, "themes")
	settings.CurrentThemeDir = filepath.Join(settings.ThemesDir, settings.CurrentTheme)

	if !utils.FileExists(settings.CurrentThemeDir) {
		log.Println("Current theme not found (", settings.CurrentTheme, "). Loading default theme")

		if settings.CurrentTheme == DefaultTheme {
			log.Println("Default theme not found")
		}

		settings.CurrentTheme = DefaultTheme
	}

	settings.CurrentThemeStylePath = filepath.Join(settings.CurrentThemeDir, "style.css")
	settings.CurrentThemeConfigPath = filepath.Join(settings.CurrentThemeDir, settings.CurrentTheme+".jsonc")

	themeConfig := cfg.ReadTheme(settings.CurrentThemeConfigPath, settings.Config)
	if themeConfig != nil {
		settings.Blur = themeConfig.Blur
		settings.Spacing = themeConfig.Spacing
		settings.PreviewStyle = themeConfig.PreviewStyle
	}

	if settings.Blur == "true" {
		enableBlur()
		ipc.AddEventListener("configreloaded", func(event string) {
			go enableBlur()
		}, true)
	}

	return settings, nil
}

func setConfigDir(mode string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Home dir: " + err.Error())
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Exe dir: " + err.Error())
	}

	exeDir := filepath.Dir(exePath)

	runModes := map[string]func() string{
		"normal": func() string {
			return filepath.Join(homeDir, ".config/hypr-dock")
		},
		"dev": func() string {
			return filepath.Join(filepath.Dir(exeDir), "configs")
		},
	}
	return runModes[mode]()
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func enableBlur() {
	ipc.Hyprctl("keyword layerrule blur,hypr-dock")
	ipc.Hyprctl("keyword layerrule ignorealpha 0.1,hypr-dock")

	ipc.Hyprctl("keyword layerrule blur,dock-popup")
	ipc.Hyprctl("keyword layerrule ignorealpha 0.1,dock-popup")
}
