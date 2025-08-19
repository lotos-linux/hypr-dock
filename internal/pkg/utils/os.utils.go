package utils

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
)
func Launch(command string) {
    cmd := exec.Command("sh", "-c", command)
    log.Printf("Launching command: %s\n", command)
    
    if err := cmd.Start(); err != nil {
        log.Printf("Unable to launch command: %s, error: %v\n", command, err)
    }
}

func LoadTextFile(path string) ([]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(bytes), "\n")
	var output []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			output = append(output, line)
		}

	}
	return output, nil
}

func TempDir() string {
	if os.Getenv("TMPDIR") != "" {
		return os.Getenv("TMPDIR")
	} else if os.Getenv("TEMP") != "" {
		return os.Getenv("TEMP")
	} else if os.Getenv("TMP") != "" {
		return os.Getenv("TMP")
	}
	return "/tmp"
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
