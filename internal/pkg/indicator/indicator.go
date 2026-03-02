package indicator

import (
	"errors"
	"fmt"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/settings"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

// IndicatorFile represents an indicator image file
type IndicatorFile struct {
	Number    int    // Numeric prefix of the file (e.g. 0 from "0.svg")
	Extension string // File extension including dot (e.g. ".svg")
	FullName  string // Complete filename (e.g. "0.svg")
}

// New creates a new indicator image based on instances count
func New(instances int, settings *settings.Settings, parent gtk.IWidget) (*gtk.Image, error) {
	available, err := GetAvailable(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to get indicators: %w", err)
	}
	if len(available) < 2 {
		return nil, errors.New("at least 2 indicator files required (0.* and *.*)")
	}
	if available[0].Number != 0 {
		return nil, errors.New("at least 2 indicator files required (0.* and *.*)")
	}

	selected := selectIndicatorFile(instances, available)
	path := filepath.Join(settings.ThemeDir, "point", selected.FullName)
	return utils.CreateImageWidthScale(path, settings.IconSize, parent, 0.56)
}

// selectIndicatorFile chooses the appropriate indicator file based on instances count
func selectIndicatorFile(instances int, files []IndicatorFile) IndicatorFile {
	selected := files[0]
	for _, file := range files {
		if file.Number > instances {
			break
		}
		selected = file
	}
	return selected
}

// GetAvailable returns all valid indicator files sorted by their numeric value
func GetAvailable(settings *settings.Settings) ([]IndicatorFile, error) {
	dirPath := filepath.Join(settings.ThemeDir, "point")

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("indicator directory not found: %w", err)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read indicator directory: %w", err)
	}

	var files []IndicatorFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file, ok := parseIndicatorFile(entry.Name())
		if ok {
			files = append(files, file)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Number < files[j].Number
	})

	return files, nil
}

// parseIndicatorFile attempts to parse a filename into an IndicatorFile
func parseIndicatorFile(name string) (IndicatorFile, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	if !isSupportedExtension(ext) {
		return IndicatorFile{}, false
	}

	baseName := name[:len(name)-len(ext)]
	num, err := strconv.Atoi(baseName)
	if err != nil {
		return IndicatorFile{}, false
	}

	return IndicatorFile{
		Number:    num,
		Extension: ext,
		FullName:  name,
	}, true
}

// isSupportedExtension checks if the extension is valid for indicator files
func isSupportedExtension(ext string) bool {
	supported := []string{".svg", ".jpg", ".png", ".webp"}
	return slices.Contains(supported, ext)
}
