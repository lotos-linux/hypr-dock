package desktop

import (
	"errors"
	"hypr-dock/internal/pkg/utils"
	"strings"
)

type Desktop2 map[string]map[string]string

func Parse(path string) (*Desktop2, error) {
	lines, err := utils.LoadTextFile(path)
	if err != nil {
		return nil, err
	}

	result := make(Desktop2)
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

	return &result, nil
}

type Action struct {
	Name string
	Exec string
	Icon string
}

func GetAppActions(className string) ([]*Action, error) {
	appDataPtr, err := Parse(SearchDesktopFile(className))
	if err != nil {
		return nil, err
	}

	appData := *appDataPtr
	general, exist := appData["desktop entry"]
	if !exist {
		return nil, errors.New("'desktop entry' section not found")
	}

	actionsStr, exist := general["actions"]
	if !exist {
		return nil, errors.New("'actions' field not found in desktop entry")
	}

	actionsList := strings.Split(actionsStr, ";")
	var actionsRes []*Action

	for _, action := range actionsList {
		if action == "" {
			continue
		}

		key := "desktop action " + action
		actionGroup, exist := appData[key]
		if !exist {
			continue
		}

		name, exist := actionGroup["name"]
		if !exist {
			name = action
		}

		exec, exist := actionGroup["exec"]
		if !exist {
			continue
		}

		icon, exist := actionGroup["icon"]
		if !exist {
			icon = ""
		}

		actionsRes = append(actionsRes, &Action{
			Name: name,
			Exec: exec,
			Icon: icon,
		})
	}

	if len(actionsRes) == 0 {
		return nil, errors.New("no valid actions found")
	}

	return actionsRes, nil
}
