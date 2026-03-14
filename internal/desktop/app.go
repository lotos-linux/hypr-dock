package desktop

import (
	"fmt"
	"hypr-dock/pkg/ini"
	"strings"

	"github.com/pkg/errors"
)

type App struct {
	name         map[string]string
	comment      map[string]string
	icon         string
	exec         string
	singleWindow bool
	actions      []Action
	raw          map[string]map[string]string

	lang string
}

type Action struct {
	name map[string]string
	exec string
	icon string

	lang string
}

func New(className string, lang ...string) (*App, error) {
	var locale string
	if len(lang) == 1 {
		locale = lang[0]
	}

	errData := &App{
		name:         map[string]string{"": className},
		comment:      map[string]string{"": className},
		icon:         className,
		exec:         className,
		singleWindow: false,
		actions:      []Action{},
		raw:          make(map[string]map[string]string),

		lang: locale,
	}

	if className == "" {
		return errData, errors.New("className empty")
	}

	file := SearchDesktopFile(className)

	raw, err := ini.GetMap(file, "Desktop Entry")
	if err != nil {
		return errData, err
	}

	general, exist := raw["Desktop Entry"]
	if !exist {
		return errData, fmt.Errorf("section \"%v\" not found in %s", "Desktop Entry", file)
	}

	name, ok := GetAllLocales(general, "Name")
	if !ok {
		name = errData.name
	}

	comment, ok := GetAllLocales(general, "Comment")
	if !ok {
		comment = errData.comment
	}

	icon, exist := general["Icon"]
	if !exist {
		icon = errData.icon
	}

	exec, exist := general["Exec"]
	if !exist {
		exec = errData.exec
	}

	singleWindowStr, exist := general["SingleMainWindow"]
	singleWindow := exist && singleWindowStr == "true"

	actions := GetActions(raw, locale)

	return &App{
		name:         name,
		comment:      comment,
		icon:         icon,
		exec:         exec,
		singleWindow: singleWindow,
		actions:      actions,
		raw:          raw,

		lang: locale,
	}, nil
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

func (a *App) GetName() string {
	return GetLocalizedValue(a.name, a.lang)
}

func (a *App) GetComment() string {
	return GetLocalizedValue(a.comment, a.lang)
}

func (a *App) Run() error {
	return run(a.exec)
}

func (a *Action) Run() error {
	return run(a.exec)
}

func run(cmd string) error {
	clean, err := CleanExec(cmd)
	if err != nil {
		return err
	}

	return Launch(clean)
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
