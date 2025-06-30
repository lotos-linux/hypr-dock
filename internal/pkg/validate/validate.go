package validate

import (
	"log"
	"slices"
)

func Preview(value string, runtime bool) bool {
	validList := []string{"live", "static", "none"}
	return Allowed("Preview", value, validList, runtime, true)
}

func Layer(value string, runtime bool) bool {
	validList := []string{"auto", "exclusive-top", "exclusive-bottom", "background", "bottom", "top", "overlay"}
	return Allowed("Layer", value, validList, runtime, true)
}

func Position(value string, runtime bool) bool {
	validList := []string{"left", "right", "top", "bottom"}
	return Allowed("Position", value, validList, runtime, true)
}

func Blur(value string, runtime bool) bool {
	validList := []string{"true", "false"}
	return Allowed("Blur", value, validList, runtime, false)
}

func SystemGapUsed(value string, runtime bool) bool {
	validList := []string{"true", "false"}
	return Allowed("SystemGapUsed", value, validList, runtime, true)
}

func Allowed[T comparable](key string, value T, validList []T, runtime bool, logs bool) bool {
	if slices.Contains(validList, value) {
		return true
	}

	if !logs {
		return false
	}

	if runtime {
		log.Printf("%s \"%v\" is incorrect or empty", key, value)
		return false
	}

	log.Printf("%s \"%v\" is incorrect or empty. Default value will be used", key, value)
	return false
}
