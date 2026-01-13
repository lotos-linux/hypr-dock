package switcher

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/tailscale/hujson"
)

type Config struct {
	WidthPercent    int  `json:"widthPercent"`    // 0-100, default 30
	HeightPercent   int  `json:"heightPercent"`   // 0-100, default 50
	FontSize        int  `json:"fontSize"`        // Default 32
	PreviewWidth    int  `json:"previewWidth"`    // Default 300
	ShowAllMonitors bool `json:"showAllMonitors"` // Default true
	CycleWorkspaces bool `json:"cycleWorkspaces"`
	IconSize        int  `json:"iconSize"` // Default 0 (Auto)
}

func GetDefaultConfig() Config {
	return Config{
		WidthPercent:    100,
		HeightPercent:   60,
		FontSize:        20,
		PreviewWidth:    400,
		ShowAllMonitors: false,
		CycleWorkspaces: true,
		IconSize:        0,
	}
}

func LoadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return GetDefaultConfig()
	}

	configPath := filepath.Join(home, ".config/hypr-dock/switcher.jsonc")
	// log.Printf("Loading switcher config from: %s", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		// log.Printf("Failed to read config: %v", err)
		return GetDefaultConfig()
	}

	// Standardize JSONC (remove comments)
	std, err := hujson.Standardize(data)
	if err != nil {
		// log.Printf("Failed to standardize JSONC: %v", err)
		return GetDefaultConfig()
	}

	var cfg Config
	if err := json.Unmarshal(std, &cfg); err != nil {
		// log.Printf("Failed to unmarshal config: %v", err)
		return GetDefaultConfig()
	}

	// log.Printf("Switcher config loaded: %+v", cfg)

	// Validate/Defaults
	if cfg.FontSize < 10 {
		cfg.FontSize = 20
	}
	if cfg.WidthPercent < 10 || cfg.WidthPercent > 100 {
		cfg.WidthPercent = 100
	}
	if cfg.HeightPercent < 10 || cfg.HeightPercent > 100 {
		cfg.HeightPercent = 60
	}
	if cfg.PreviewWidth < 50 {
		cfg.PreviewWidth = 400
	}

	return cfg
}
