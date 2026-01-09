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
	envs := []string{
		"TMPDIR",
		"TEMP",
		"TMP",
	}

	for _, env := range envs {
		val := os.Getenv(env)
		if val != "" {
			return val
		}
	}
	return "/tmp"
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
