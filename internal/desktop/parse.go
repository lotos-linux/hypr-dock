package desktop

import (
	"strings"
)

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

func GetActions(raw map[string]map[string]string, lang ...string) []Action {
	var actionsRes []Action
	var locale string
	if len(lang) == 1 {
		locale = lang[0]
	}

	general, exist := raw["Desktop Entry"]
	if !exist {
		return actionsRes
	}

	actionsStr, exist := general["Actions"]
	if !exist {
		return actionsRes
	}

	actionsList := strings.Split(actionsStr, ";")

	for _, actionName := range actionsList {
		actionName = strings.TrimSpace(actionName)
		if actionName == "" {
			continue
		}

		key := "Desktop Action " + actionName
		actionGroup, exist := raw[key]
		if !exist {
			continue
		}

		name, ok := GetAllLocales(actionGroup, "Name")
		if !ok {
			name = map[string]string{"": actionName}
		}

		exec, exist := actionGroup["Exec"]
		if !exist {
			continue
		}

		icon, exist := actionGroup["Icon"]
		if !exist {
			icon = ""
		}

		actionsRes = append(actionsRes, Action{
			name: name,
			exec: exec,
			icon: icon,

			lang: locale,
		})
	}

	return actionsRes
}
