package hyprEvents

import (
	"fmt"
	"log"
	"strings"

	"github.com/gotk3/gotk3/glib"

	"hypr-dock/internal/app"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
)

func Init(appState *state.State) {
	ipc.AddEventListener("windowtitlev2", func(event string) {
		windowTitleHandler(event, appState)
	}, true)

	ipc.AddEventListener("openwindow", func(event string) {
		openwindowHandler(event, appState)
	}, true)

	ipc.AddEventListener("closewindow", func(event string) {
		closewindowHandler(event, appState)
	}, true)

	ipc.AddEventListener("activespecial", func(event string) {
		activatespecialHandler(event, appState)
	}, true)

	go ipc.InitHyprEvents()
}

func windowTitleHandler(event string, appState *state.State) {
	data := eventHandler(event, 2)
	address := "0x" + strings.TrimSpace(data[0])
	go app.ChangeWindowTitle(address, data[1], appState)
}

func activatespecialHandler(event string, appState *state.State) {
	data := eventHandler(event, 2)
	log.Printf("Received activespecial event: %v", data)

	if data[0] == "special:special" {
		log.Println("Special workspace activated")
		appState.SetSpecial(true)
	} else {
		log.Println("Special workspace deactivated")
		appState.SetSpecial(false)
	}
}

func openwindowHandler(event string, appState *state.State) {
	data := eventHandler(event, 4)
	address := "0x" + strings.TrimSpace(data[0])
	windowClient, err := ipc.SearchClientByAddress(address)
	if err != nil {
		fmt.Println(err)
	} else {
		glib.IdleAdd(func() {
			app.InitNewItemInIPC(windowClient, appState)
		})
	}
}

func closewindowHandler(event string, appState *state.State) {
	data := eventHandler(event, 1)
	address := "0x" + strings.TrimSpace(data[0])
	glib.IdleAdd(func() {
		app.RemoveApp(address, appState)
	})
}

func eventHandler(event string, n int) []string {
	parts := strings.SplitN(event, ">>", 2)
	dataClast := strings.TrimSpace(parts[1])
	dataParts := strings.SplitN(dataClast, ",", n)

	for i := range dataParts {
		dataParts[i] = strings.TrimSpace(dataParts[i])
	}

	return dataParts
}

