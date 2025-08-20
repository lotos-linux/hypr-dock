package desktop

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Desktop struct {
	Name         string
	Icon         string
	Exec         string
	SingleWindow bool
}

var desktopDirs = GetAppDirs()

func New(className string) *Desktop {
	errData := &Desktop{
		Name:         "Untitle",
		Icon:         "",
		Exec:         "",
		SingleWindow: false,
	}

	allData, err := Parse(SearchDesktopFile(className))
	if err != nil {
		return errData
	}

	appData := *allData
	general, exist := appData["desktop entry"]
	if !exist {
		return errData
	}

	name, exist := general["name"]
	if !exist {
		name = errData.Name
	}

	icon, exist := general["icon"]
	if !exist {
		icon = ""
	}

	exec, exist := general["exec"]
	if !exist {
		exec = ""
	}

	singleWindowStr, exist := general["singlemainwindow"]
	if !exist {
		singleWindowStr = "false"
	}

	singleWindow := singleWindowStr == "true"

	return &Desktop{
		Name:         name,
		Icon:         icon,
		Exec:         exec,
		SingleWindow: singleWindow,
	}
}

func GetSingleWindow(allData []string) bool {
	val := GetDesktopOption(allData, "SingleMainWindow")
	return val == "true"
}

func SearchDesktopFile(className string) string {
	for _, appDir := range desktopDirs {
		desktopFile := className + ".desktop"
		_, err := os.Stat(filepath.Join(appDir, desktopFile))
		if err == nil {
			return filepath.Join(appDir, desktopFile)
		}

		// If file non found
		files, _ := os.ReadDir(appDir)
		for _, file := range files {
			fileName := file.Name()

			// "krita" > "org.kde.krita.desktop" / "lutris" > "net.lutris.Lutris.desktop"
			if strings.Count(fileName, ".") > 1 && strings.Contains(fileName, className) {
				return filepath.Join(appDir, fileName)
			}
			// "VirtualBox Manager" > "virtualbox.desktop"
			if fileName == strings.Split(strings.ToLower(className), " ")[0]+".desktop" {
				return filepath.Join(appDir, fileName)
			}
		}
	}

	return ""
}

func GetDesktopOption(allData []string, option string) string {
	for lineIndex := range len(allData) {
		line := allData[lineIndex]
		if strings.HasPrefix(line, option+"=") {
			optionValue := strings.Split(line, "=")[1]
			return optionValue
		}
	}
	return ""
}

func GetAppDirs() []string {
	var dirs []string
	xdgDataDirs := ""

	home := os.Getenv("HOME")
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if os.Getenv("XDG_DATA_DIRS") != "" {
		xdgDataDirs = os.Getenv("XDG_DATA_DIRS")
	} else {
		xdgDataDirs = "/usr/local/share/:/usr/share/"
	}
	if xdgDataHome != "" {
		dirs = append(dirs, filepath.Join(xdgDataHome, "applications"))
	} else if home != "" {
		dirs = append(dirs, filepath.Join(home, ".local/share/applications"))
	}
	for _, d := range strings.Split(xdgDataDirs, ":") {
		dirs = append(dirs, filepath.Join(d, "applications"))
	}
	flatpakDirs := []string{filepath.Join(home, ".local/share/flatpak/exports/share/applications"),
		"/var/lib/flatpak/exports/share/applications"}

	for _, d := range flatpakDirs {
		if !slices.Contains(dirs, d) {
			dirs = append(dirs, d)
		}
	}
	return dirs
}
