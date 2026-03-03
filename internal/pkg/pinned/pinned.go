package pinned

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func Open(path string) ([]string, error) {
	err := createFile(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func Save(path string, newList []string) error {
	err := createFile(path)
	if err != nil {
		return err
	}

	var content strings.Builder
	for _, line := range newList {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" {
			content.WriteString(cleanLine)
			content.WriteByte('\n')
		}
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content.String()), 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func createFile(path string) error {
	var err error

	dir := filepath.Dir(path)
	_, err = os.Stat(dir)

	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		err := os.WriteFile(path, []byte{}, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
