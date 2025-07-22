package detectzone

import (
	"hypr-dock/internal/state"
	"log"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func Init(appState *state.State) {
	window := appState.GetWindow()

	detectArea, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("InitDetectArea(), gtk.WindowNew() | ", err)
	}
	detectArea.SetName("detect")
	detectArea.SetSizeRequest(-1, 1)

	layershell.InitForWindow(detectArea)
	layershell.SetNamespace(detectArea, "dock-detect")
	layershell.SetLayer(detectArea, layershell.LAYER_SHELL_LAYER_TOP)
	selectEdges(detectArea, appState)

	detectArea.Connect("enter-notify-event", func(detectWindow *gtk.Window, e *gdk.Event) {
		timer := appState.GetDockHideTimer()
		timer.Stop()

		go func() {
			layershell.SetLayer(window, layershell.LAYER_SHELL_LAYER_TOP)
		}()
	})

	detectArea.Connect("leave-notify-event", func(detectWindow *gtk.Window, e *gdk.Event) {
		timer := appState.GetDockHideTimer()

		timer.Run(appState.Settings.AutoHideDeley, func() {
			layershell.SetLayer(window, layershell.LAYER_SHELL_LAYER_BOTTOM)
		})
	})

	detectArea.ShowAll()
	appState.SetDetectArea(detectArea)
}

func selectEdges(window *gtk.Window, appState *state.State) {
	settings := appState.GetSettings()

	switch settings.Position {
	case "left":
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_LEFT, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_TOP, true)
		layershell.SetMargin(window, layershell.LAYER_SHELL_EDGE_LEFT, 0)
	case "top":
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_RIGHT, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_LEFT, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_TOP, true)
		layershell.SetMargin(window, layershell.LAYER_SHELL_EDGE_TOP, 0)
	case "right":
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_RIGHT, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_TOP, true)
		layershell.SetMargin(window, layershell.LAYER_SHELL_EDGE_RIGHT, 0)
	case "bottom":
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_LEFT, true)
		layershell.SetAnchor(window, layershell.LAYER_SHELL_EDGE_RIGHT, true)
		layershell.SetMargin(window, layershell.LAYER_SHELL_EDGE_BOTTOM, 0)
	}
}
