package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/allan-simon/go-singleinstance"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/app"
	"hypr-dock/internal/hypr/hyprEvents"
	"hypr-dock/internal/layering"
	"hypr-dock/internal/pkg/flags"
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

	// flags
	flags := flags.Get()

	logger := utils.СreateLogger(flags.LogLevel)

	// window build
	settings, err := settings.Init(flags, logger)
	if err != nil {
		logger.Error("Settings init error:", "err", err)
	}

	gtk.Init(nil)

	appState := state.New(settings)
	appState.SetLogger(logger)

	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		logger.Error("Unable to create window:", "err", err)
		os.Exit(2)
	}
	appState.SetWindow(window)

	window.SetTitle("hypr-dock")

	layerctl := layering.NewInit(window, settings)
	appState.SetLayerctl(layerctl)

	err = utils.AddCssProvider(settings.ThemeStyle)
	if err != nil {
		logger.Warn("CSS file not found, the default GTK theme is running!", "err", err)
	}

	app := app.BuildApp(appState)

	window.Add(app)
	window.Connect("destroy", func() { gtk.MainQuit() })
	window.ShowAll()

	// post
	hyprEvents.Init(appState)

	// end
	gtk.Main()
}
