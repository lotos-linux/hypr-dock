package switcher

import (
	"hypr-dock/pkg/ini"
	"log"

	"github.com/hashicorp/go-hclog"
)

type Config struct {
	WidthPercent    int  `def:"100" max:"100"`
	HeightPercent   int  `def:"60" max:"100"`
	FontSize        int  `def:"32"`
	PreviewWidth    int  `def:"300"`
	ShowAllMonitors bool `def:"false"`
	CycleWorkspaces bool `def:"true"`
	IconSize        int  `def:"0"`
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

func LoadConfig(path string) Config {
	cfg := ini.New(path, hclog.Default())

	var conf Config
	_, err := cfg.ParseSection(&conf, "General")
	if err != nil {
		return GetDefaultConfig()
	}

	log.Println(conf.FontSize)

	return conf
}
