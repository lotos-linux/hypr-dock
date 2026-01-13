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
	"hypr-dock/internal/switcher"
)

func main() {
	// Handle flags
	args := os.Args[1:]

	// Check if run as "hypr-alttab"
	exe, _ := os.Executable()
	if len(os.Args) > 0 && (os.Args[0] == "hypr-alttab" || contains(exe, "hypr-alttab")) {
		switcher.Run()
		return
	}

	if len(args) > 0 && args[0] == "--switcher" {
		switcher.Run()
		return
	}

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
	settings, err := settings.Init()
	if err != nil {
		log.Println("Settings init error: ", err)
	}

	gtk.Init(nil)

	appState := state.New(settings)

	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	appState.SetWindow(window)

	window.SetTitle("hypr-dock")

	layering.SetWindowProperty(appState)

	err = utils.AddCssProvider(settings.CurrentThemeStylePath)
	if err != nil {
		log.Println("CSS file not found, the default GTK theme is running!\n", err)
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

func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if hasPrefix(s[i:], substr) {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
