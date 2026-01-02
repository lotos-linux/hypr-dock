package desktop

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func SearchDesktopFile(className string) string {
	for _, appDir := range GetAppDirs() {
		desktopFile := className + ".desktop"
		_, err := os.Stat(filepath.Join(appDir, desktopFile))
		if err == nil {
			return filepath.Join(appDir, desktopFile)
		}

		// If file non found
		files, _ := os.ReadDir(appDir)

		// "krita" > "org.kde.krita.desktop" / "lutris" > "net.lutris.Lutris.desktop"
		for _, file := range files {
			if strings.Count(file.Name(), ".") > 1 && strings.Contains(file.Name(), className) {
				return filepath.Join(appDir, file.Name())
			}

		}

		// "VirtualBox Manager" > "virtualbox.desktop"
		for _, file := range files {
			if file.Name() == strings.Split(strings.ToLower(className), " ")[0]+".desktop" {
				return filepath.Join(appDir, file.Name())
			}
		}

		// "GitHub Desktop" > "github-desktop.desktop"
		for _, file := range files {
			fileName := file.Name()
			fileName = strings.ToLower(fileName)
			classNameLower := strings.ToLower(className)
			classNameLower = strings.ReplaceAll(classNameLower, " ", "-")

			if fileName == classNameLower+".desktop" {
				return filepath.Join(appDir, file.Name())
			}
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
