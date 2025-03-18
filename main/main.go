package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/allan-simon/go-singleinstance"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/app"
	"hypr-dock/internal/hypr/hyprEvents"
	"hypr-dock/internal/layering"
	"hypr-dock/internal/pkg/signals"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/settings"
	"hypr-dock/internal/state"
)

func main() {
	signals.Handler()

	lockFilePath := fmt.Sprintf("%s/hypr-dock-%s.lock", utils.TempDir(), os.Getenv("USER"))
	lockFile, err := singleinstance.CreateLockFile(lockFilePath)
	if err != nil {
		file, err := utils.LoadTextFile(lockFilePath)
		if err == nil {
			pidStr := file[0]
			pidInt, _ := strconv.Atoi(pidStr)
			syscall.Kill(pidInt, syscall.SIGUSR1)
		}
		os.Exit(0)
	}
	defer lockFile.Close()

	// window build
	settings.Init()
	gtk.Init(nil)

	appState := state.New()

	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Println("Unable to create window:", err)
	}
	appState.SetWindow(window)

	window.SetTitle("hypr-dock")

	orientation, edge := layering.SetWindowProperty(window, appState)
	appState.SetOrientation(orientation)

	err = utils.AddCssProvider(settings.CurrentThemeStylePath)
	if err != nil {
		log.Println("CSS file not found, the default GTK theme is running!\n", err)
	}

	app := app.BuildApp(orientation, appState)

	window.Add(app)
	window.Connect("destroy", func() { gtk.MainQuit() })
	window.ShowAll()

	// post
	if settings.Get().Layer == "auto" {
		layering.InitDetectArea(edge, appState)
	}

	hyprEvents.Init(appState)

	// end
	gtk.Main()
}
