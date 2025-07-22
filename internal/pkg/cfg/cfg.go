package cfg

import (
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/pkg/validate"
	"log"
	"os"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/tailscale/hujson"
)

type Config struct {
	CurrentTheme    string
	IconSize        int
	Layer           string
	Position        string
	Blur            string
	Spacing         int
	AutoHideDeley   int
	SystemGapUsed   string
	Margin          int
	ContextPos      int
	Preview         string
	PreviewAdvanced struct {
		FPS        int
		BufferSize int
		ShowDelay  int
		HideDelay  int
		MoveDelay  int
	}
	PreviewStyle struct {
		Size         int
		BorderRadius int
		Padding      int
	}
}

type ThemeConfig struct {
	Blur         string
	Spacing      int
	PreviewStyle struct {
		Size         int
		BorderRadius int
		Padding      int
	}
}

type ItemList struct {
	Pinned []string
}

func GetDefaultConfig() Config {
	return Config{
		CurrentTheme:  "lotos",
		IconSize:      21,
		Layer:         "auto",
		Position:      "bottom",
		Blur:          "true",
		Spacing:       8,
		SystemGapUsed: "true",
		AutoHideDeley: 400,
		Margin:        8,
		ContextPos:    5,
		Preview:       "none",
		PreviewAdvanced: struct {
			FPS        int
			BufferSize int
			ShowDelay  int
			HideDelay  int
			MoveDelay  int
		}{
			FPS:        30,
			BufferSize: 5,
			ShowDelay:  600,
			HideDelay:  300,
			MoveDelay:  200,
		},
		PreviewStyle: struct {
			Size         int
			BorderRadius int
			Padding      int
		}{
			Size:         120,
			BorderRadius: 0,
			Padding:      10,
		},
	}
}

func ReadConfig(jsoncFile string, themesDir string) Config {
	// Read jsonc
	config := Config{}
	err := ReadJsonc(jsoncFile, &config)
	if err != nil {
		log.Println(err)
		log.Println("Load default config")
		return GetDefaultConfig()
	}

	// Set default values ​​if not specified
	if config.CurrentTheme == "" {
		config.CurrentTheme = GetDefaultConfig().CurrentTheme
		log.Println("The theme is not set, the default theme is currently used - \"lotos\"")
	}

	if !validate.Layer(config.Layer, false) {
		config.Layer = GetDefaultConfig().Layer
	}

	if !validate.Position(config.Position, false) {
		config.Position = GetDefaultConfig().Position
	}

	if !validate.Blur(config.Blur, false) {
		config.Blur = GetDefaultConfig().Blur
	}

	if !validate.SystemGapUsed(config.SystemGapUsed, false) {
		config.SystemGapUsed = GetDefaultConfig().SystemGapUsed
	}

	if !validate.Preview(config.Preview, false) {
		config.Preview = GetDefaultConfig().Preview
	}

	if config.PreviewAdvanced.FPS == 0 {
		config.PreviewAdvanced.FPS = GetDefaultConfig().PreviewAdvanced.FPS
	}

	if config.PreviewAdvanced.BufferSize < 1 || config.PreviewAdvanced.BufferSize > 20 {
		config.PreviewAdvanced.BufferSize = GetDefaultConfig().PreviewAdvanced.BufferSize
	}

	if config.Spacing < 1 {
		config.Spacing = GetDefaultConfig().Spacing
	}

	if config.IconSize < 1 {
		config.IconSize = GetDefaultConfig().IconSize
	}

	return config
}

func ReadTheme(jsoncFile string, config Config) *ThemeConfig {
	// Read jsonc
	themeConfig := ThemeConfig{}
	err := ReadJsonc(jsoncFile, &themeConfig)
	if err != nil {
		log.Println(err)
		log.Println("Load default config")
		return nil
	}

	// Set default values ​​if not specified
	if !validate.Blur(config.Blur, false) {
		themeConfig.Blur = config.Blur
	}

	if themeConfig.Spacing < 0 {
		themeConfig.Spacing = config.Spacing
	}

	if themeConfig.PreviewStyle.Size < 20 {
		themeConfig.PreviewStyle.Size = GetDefaultConfig().PreviewStyle.Size
	}

	return &themeConfig
}

func ReadItemList(jsonFile string) []string {
	itemList := ItemList{}

	if !utils.FileExists(jsonFile) {
		itemList.Pinned = CreateEmptyPinnedFile(jsonFile)
		return itemList.Pinned
	}

	err := ReadJson(jsonFile, &itemList)
	if err != nil {
		log.Fatal(err)
	}

	return itemList.Pinned
}

func ReadJsonc(jsoncFile string, v interface{}) error {
	file, err := os.ReadFile(jsoncFile)
	if err != nil {
		return errors.Wrapf(err, "file %q not found", jsoncFile)
	}

	// Парсим JSONC
	standardized, err := hujson.Standardize(file)
	if err != nil {
		return errors.Wrapf(err, "failed to standardize JSONC")
	}

	if err := json.Unmarshal(standardized, &v); err != nil {
		return errors.Wrapf(err, "file %q has a syntax error", jsoncFile)
	}

	return nil
}

func ChangeJsonPinnedApps(apps []string, jsonFile string) error {
	itemList := ItemList{
		Pinned: apps,
	}

	if err := WriteItemList(jsonFile, itemList); err != nil {
		log.Println("Error", jsonFile, "writing: ", err)
		return err
	}

	return nil
}

func ReadJson(jsonFile string, v interface{}) error {
	file, err := os.Open(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&v); err != nil {
		return err
	}

	return nil
}

func CreateEmptyPinnedFile(jsonFile string) []string {
	initialData := ItemList{
		Pinned: []string{},
	}

	if err := WriteItemList(jsonFile, initialData); err != nil {
		log.Fatalf("Failed to create file %q: %v", jsonFile, err)
		return nil
	}

	return initialData.Pinned
}

func WriteItemList(jsonFile string, data ItemList) error {
	file, err := os.Create(jsonFile)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", jsonFile)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return errors.Wrapf(err, "failed to encode data to file %q", jsonFile)
	}

	return nil
}
