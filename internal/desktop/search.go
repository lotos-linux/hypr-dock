package desktop

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func SearchDesktopFile(className string) string {
	for _, appDir := range GetAppDirs() {
		_, err := os.Stat(filepath.Join(appDir, className+".desktop"))
		if err == nil {
			return filepath.Join(appDir, className+".desktop")
		}

		// If file non found
		files, _ := os.ReadDir(appDir)

		// "krita" > "org.kde.krita.desktop" / "lutris" > "net.lutris.Lutris.desktop"
		for _, file := range files {
			if strings.Count(file.Name(), ".") > 1 && strings.Contains(strings.ToLower(file.Name()), className) {
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

		// Chrome/Chromium webapp: "chrome-messenger.com__-Default" > "Messenger.desktop" (by martonbtoth)
		if strings.HasPrefix(className, "chrome-") || strings.HasPrefix(className, "chromium-") {
			// Extract domain from class name (e.g., "chrome-messenger.com__-Default" -> "messenger.com")
			parts := strings.SplitN(className, "-", 2)
			if len(parts) == 2 {
				domain := strings.Split(parts[1], "__")[0] // Remove "__-Default" suffix
				domain = strings.TrimSuffix(domain, "-")
				domainParts := strings.Split(domain, ".")
				if len(domainParts) > 0 {
					// Try matching by domain name (e.g., "messenger" from "messenger.com")
					baseName := domainParts[0]
					for _, file := range files {
						fileName := file.Name()
						fileNameLower := strings.ToLower(fileName)
						if strings.Contains(fileNameLower, strings.ToLower(baseName)) && strings.HasSuffix(fileNameLower, ".desktop") {
							return filepath.Join(appDir, fileName)
						}
					}
				}
			}
		}
	}

	path, exist := GetFiles()[className]
	if exist {
		return path
	}

	return ""
}

var (
	dirs  []string
	dOnce sync.Once
)

func GetAppDirs() []string {
	dOnce.Do(func() {
		dirs = newAppDirs()
	})
	return dirs
}

func newAppDirs() []string {
	home, _ := os.UserHomeDir()

	return ProcessDirectories(append([]string{
		// dock custom apps
		filepath.Join(home, ".local/share/hypr-dock"),

		// user local apps
		os.Getenv("XDG_DATA_HOME"),
		filepath.Join(home, ".local/share"),

		// flatpak
		filepath.Join(home, ".local/share/flatpak/exports/share"),
		"/var/lib/flatpak/exports/share",

		// system
		"/usr/local/share",
		"/usr/share",

		// xdg
	}, strings.Split(os.Getenv("XDG_DATA_DIRS"), ":")...))
}

func ProcessDirectories(paths []string) []string {
	uniquePaths := make(map[string]bool)
	var result []string

	for _, path := range paths {
		if path == "" {
			continue
		}

		path = filepath.Clean(path)
		path = filepath.Join(path, "applications")

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
