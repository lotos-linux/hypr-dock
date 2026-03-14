package utils

import (
	"regexp"
	"strings"

	"github.com/hashicorp/go-hclog"
)

func GetSingleValue[K comparable, V any](m map[K]V) (V, bool) {
	for _, v := range m {
		return v, true
	}
	var zero V
	return zero, false
}

func СreateLogger(logLevel string) hclog.Logger {
	level := hclog.LevelFromString(logLevel)

	if level == hclog.NoLevel {
		level = hclog.Info
	}

	return hclog.New(&hclog.LoggerOptions{
		Name:  "hypr-dock",
		Level: level,
		Color: hclog.AutoColor,
	})
}

func NormaliseTitle(title string) string {
	re := regexp.MustCompile(`^[a-zA-Z-]+`)
	firstWord := re.FindString(title)

	return strings.ToLower(firstWord)
}
