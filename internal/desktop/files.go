package desktop

import (
	"hypr-dock/pkg/ini"
	"os"
	"path/filepath"
	"sync"
)

var (
	table map[string]string
	once  sync.Once
)

func GetFiles() map[string]string {
	once.Do(func() {
		table = newTable()
	})
	return table
}

func newTable() map[string]string {
	res := make(map[string]string)
	dirs := GetAppDirs()

	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			path := filepath.Join(dir, file.Name())

			data, err := ini.GetMap(path, "Desktop Entry")
			if err != nil {
				continue
			}

			general, exist := data["Desktop Entry"]
			if !exist {
				continue
			}

			className, exist := general["StartupWMClass"]
			if !exist {
				continue
			}

			if className != "" {
				res[className] = path
			}
		}
	}

	return res
}
