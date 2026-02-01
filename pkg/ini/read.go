package ini

import (
	"os"
	"strings"
)

func GetMap(path string, initialSection string) (map[string]map[string]string, error) {
	lines, err := load(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)
	currentSection := initialSection

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "#") {
			line = strings.SplitN(line, "#", 2)[0]
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if result[currentSection] == nil {
			result[currentSection] = make(map[string]string)
		}

		result[currentSection][key] = value
	}

	return result, nil
}

func load(path string) ([]string, error) {
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
