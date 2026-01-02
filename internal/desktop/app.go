package desktop

import (
	"log"
	"strings"
)

type App struct {
	name         map[string]string
	comment      map[string]string
	icon         string
	exec         string
	singleWindow bool
	actions      []Action
	raw          map[string]map[string]string
}

type Action struct {
	name map[string]string
	exec string
	icon string
}

func New(className string) *App {
	errData := &App{
		name:         map[string]string{"": className},
		comment:      map[string]string{"": ""},
		icon:         "",
		exec:         "",
		singleWindow: false,
		actions:      []Action{},
		raw:          make(map[string]map[string]string),
	}

	raw, err := Parse(SearchDesktopFile(className))
	if err != nil {
		return errData
	}

	general, exist := raw["desktop entry"]
	if !exist {
		return errData
	}

	name, ok := GetAllLocales(general, "name")
	if !ok {
		name = errData.name
	}

	comment, ok := GetAllLocales(general, "comment")
	if !ok {
		comment = errData.comment
	}

	icon, exist := general["icon"]
	if !exist {
		icon = errData.icon
	}

	exec, exist := general["exec"]
	if !exist {
		exec = errData.exec
	}

	singleWindowStr, exist := general["singlemainwindow"]
	singleWindow := exist && singleWindowStr == "true"

	actions := GetActions(raw)

	return &App{
		name:         name,
		comment:      comment,
		icon:         icon,
		exec:         exec,
		singleWindow: singleWindow,
		actions:      actions,
		raw:          raw,
	}
}

func GetLocalizedValue(values map[string]string, lang string) string {
	if name, ok := values[lang]; ok && name != "" {
		return name
	}

	if defaultName, ok := values[""]; ok && defaultName != "" {
		return defaultName
	}

	if strings.HasPrefix(lang, "en") {
		for _, fallback := range []string{"en_US", "en_GB", "en", "C"} {
			if name, ok := values[fallback]; ok && name != "" {
				return name
			}
		}
	}

	for _, name := range values {
		if name != "" {
			return name
		}
	}

	return ""
}

func (a *App) GetAllName() map[string]string {
	return a.name
}

func (a *App) GetAllComment() map[string]string {
	return a.comment
}

func (a *App) GetIcon() string {
	return a.icon
}

func (a *App) GetExec() string {
	return a.exec
}

func (a *App) GetSingleWindow() bool {
	return a.singleWindow
}

func (a *App) GetActions() []Action {
	return a.actions
}

func (a *App) GetRaw() map[string]map[string]string {
	return a.raw
}

func (a *App) GetName(lang ...string) string {
	if len(lang) < 1 {
		return GetLocalizedValue(a.name, "")
	}
	return GetLocalizedValue(a.name, lang[0])
}

func (a *App) GetComment(lang ...string) string {
	if len(lang) < 1 {
		return GetLocalizedValue(a.comment, "")
	}
	return GetLocalizedValue(a.comment, lang[0])
}

func (a *App) Run() {
	run(a.exec)
}

func (a *Action) Run() {
	run(a.exec)
}

func run(cmd string) {
	clean, err := CleanExec(cmd)
	if err != nil {
		log.Println(err)
		return
	}

	Launch(clean)
}

func (a *Action) GetAllName() map[string]string {
	return a.name
}

func (a *Action) GetExec() string {
	return a.exec
}

func (a *Action) GetIcon() string {
	return a.icon
}

func (a *Action) GetName(lang ...string) string {
	if len(lang) < 1 {
		return GetLocalizedValue(a.name, "")
	}
	return GetLocalizedValue(a.name, lang[0])
}
