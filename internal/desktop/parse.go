package desktop

import (
	"os"
	"strings"
)

func Parse(path string) (map[string]map[string]string, error) {
	lines, err := loadTextFile(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)
	currentSection := "desktop-entry"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(line[1 : len(line)-1])
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		if result[currentSection] == nil {
			result[currentSection] = make(map[string]string)
		}

		result[currentSection][key] = value
	}

	return result, nil
}

func GetAllLocales(m map[string]string, prefix string) (map[string]string, bool) {
	result := make(map[string]string)
	hasDefault := false
	hasAnyTranslation := false

	for key, value := range m {
		if key == prefix {
			result[""] = value
			hasDefault = true
			hasAnyTranslation = true
			continue
		}

		if strings.HasPrefix(key, prefix+"[") && strings.HasSuffix(key, "]") {
			lang := key[len(prefix)+1 : len(key)-1]

			if lang == "" && hasDefault {
				continue
			}

			result[lang] = value
			hasAnyTranslation = true

			if lang == "" {
				hasDefault = true
			}
		}
	}

	if !hasAnyTranslation {
		return nil, false
	}

	if !hasDefault {
		fallbackOrder := []string{"en", "en_US", "en_GB", "C", ""}

		for _, lang := range fallbackOrder {
			if value, ok := result[lang]; ok && value != "" {
				result[""] = value
				hasDefault = true
				break
			}
		}

		if !hasDefault && len(result) > 0 {
			for _, value := range result {
				if value != "" {
					result[""] = value
					break
				}
			}
		}
	}

	return result, true
}

func GetActions(raw map[string]map[string]string) []Action {
	var actionsRes []Action

	general, exist := raw["desktop entry"]
	if !exist {
		return actionsRes
	}

	actionsStr, exist := general["actions"]
	if !exist {
		return actionsRes
	}

	actionsList := strings.Split(actionsStr, ";")

	for _, actionName := range actionsList {
		actionName = strings.TrimSpace(actionName)
		if actionName == "" {
			continue
		}

		key := "desktop action " + actionName
		actionGroup, exist := raw[key]
		if !exist {
			continue
		}

		name, ok := GetAllLocales(actionGroup, "name")
		if !ok {
			name = map[string]string{"": actionName}
		}

		exec, exist := actionGroup["exec"]
		if !exist {
			continue
		}

		icon, exist := actionGroup["icon"]
		if !exist {
			icon = ""
		}

		actionsRes = append(actionsRes, Action{
			name: name,
			exec: exec,
			icon: icon,
		})
	}

	return actionsRes
}

func loadTextFile(path string) ([]string, error) {
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
