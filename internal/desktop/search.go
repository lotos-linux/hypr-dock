package desktop

import (
	"os"
	"path/filepath"
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
	home, _ := os.UserHomeDir()

	res := []string{
		// dock custom apps
		filepath.Join(home, ".local/share/hypr-dock/applications"),

		// user local apps
		filepath.Join(os.Getenv("XDG_DATA_HOME"), "applications"),
		filepath.Join(home, ".local/share/applications"),

		// flatpak
		filepath.Join(home, ".local/share/flatpak/exports/share/applications"),
		"/var/lib/flatpak/exports/share/applications",

		// system
		"/usr/local/share/",
		"/usr/share/",
	}

	// XDG standart
	for _, dir := range strings.Split(os.Getenv("XDG_DATA_DIRS"), ":") {
		res = append(res, filepath.Join(dir, "applications"))
	}

	return ProcessDirectories(res)
}

func ProcessDirectories(paths []string) []string {
	uniquePaths := make(map[string]bool)
	var result []string

	for _, path := range paths {
		if path == "" {
			continue
		}

		path = filepath.Clean(path)

		if !filepath.IsAbs(path) {
			continue
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			continue
		}

		if !fileInfo.IsDir() {
			continue
		}

		if !uniquePaths[path] {
			uniquePaths[path] = true
			result = append(result, path)
		}
	}

	return result
}
