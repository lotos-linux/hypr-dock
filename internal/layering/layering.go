package layering

import (
	detectzone "hypr-dock/internal/detectZone"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
	"strings"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func SetWindowProperty(appState *state.State) {
	window := appState.GetWindow()
	settings := appState.GetSettings()

	layershell.InitForWindow(window)
	layershell.SetNamespace(window, "hypr-dock")

	ChangePosition(settings.Position, appState)
	ChangeLayer(settings.Layer, appState)
}

func ChangeLayer(layer string, appState *state.State) {
	window := appState.GetWindow()
	if window == nil {
		return
	}

	layers := map[string]layershell.LayerShellLayerFlags{
		"background": layershell.LAYER_SHELL_LAYER_BACKGROUND,
		"bottom":     layershell.LAYER_SHELL_LAYER_BOTTOM,
		"top":        layershell.LAYER_SHELL_LAYER_TOP,
		"overlay":    layershell.LAYER_SHELL_LAYER_OVERLAY,
	}

	if layer == "auto" {
		layershell.SetLayer(window, layers["bottom"])
		layershell.SetExclusiveZone(window, 0)
		AutoLayer(appState)
		detectzone.Init(appState)
		return
	}

	DisableAutoLayer(appState)
	if strings.Contains(layer, "exclusive") {
		exLayer := strings.Split(layer, "-")[1]
		layershell.SetLayer(window, layers[exLayer])
		layershell.AutoExclusiveZoneEnable(window)
		return
	}

	layershell.SetLayer(window, layers[layer])
}

func ChangePosition(position string, appState *state.State) {
	window := appState.GetWindow()
	if window == nil {
		return
	}

	oreintations := map[string]gtk.Orientation{
		"bottom": gtk.ORIENTATION_HORIZONTAL,
		"top":    gtk.ORIENTATION_HORIZONTAL,
		"left":   gtk.ORIENTATION_VERTICAL,
		"right":  gtk.ORIENTATION_VERTICAL,
	}

	edges := map[string]layershell.LayerShellEdgeFlags{
		"bottom": layershell.LAYER_SHELL_EDGE_BOTTOM,
		"top":    layershell.LAYER_SHELL_EDGE_TOP,
		"left":   layershell.LAYER_SHELL_EDGE_LEFT,
		"right":  layershell.LAYER_SHELL_EDGE_RIGHT,
	}

	layershell.SetAnchor(window, edges[position], true)
	layershell.SetMargin(window, edges[position], 0)

	appState.SetOrientation(oreintations[position])
	appState.SetEdge(edges[position])
}

func AutoLayer(appState *state.State) {
	DisableAutoLayer(appState)
	window := appState.GetWindow()

	ipc.AddEventListener("hd>>pv-pointer-enter", func(e string) {
		appState.GetDockHideTimer().Stop()
	}, true)

	ipc.AddEventListener("hd>>pv-pointer-leave", func(e string) {
		DispathLeaveEvent(window, nil, appState)
	}, true)

	ipc.AddEventListener("hd>>focus-window", func(e string) {
		DispathLeaveEvent(window, nil, appState)
	}, true)

	enterSig := window.Connect("enter-notify-event", func(window *gtk.Window, e *gdk.Event) {
		event := gdk.EventCrossingNewFromEvent(e)
		isInWindow := event.Detail() == 3 || event.Detail() == 4

		if !isInWindow || appState.GetSpecial() {
			return
		}

		timer := appState.GetDockHideTimer()

		timer.Stop()
		layershell.SetLayer(window, layershell.LAYER_SHELL_LAYER_TOP)
	})
	appState.AddSignalHandler("enter", enterSig)

	leaveSig := window.Connect("leave-notify-event", func(window *gtk.Window, e *gdk.Event) {
		DispathLeaveEvent(window, e, appState)
	})
	appState.AddSignalHandler("leave", leaveSig)
}

func DispathLeaveEvent(window *gtk.Window, e *gdk.Event, appState *state.State) {
	isInWindow := true
	if e != nil {
		event := gdk.EventCrossingNewFromEvent(e)
		isInWindow = event.Detail() == 3 || event.Detail() == 4
	}

	if !isInWindow {
		return
	}

	timer := appState.GetDockHideTimer()

	timer.Run(appState.Settings.AutoHideDeley, func() {
		layershell.SetLayer(window, layershell.LAYER_SHELL_LAYER_BOTTOM)
	})
}

func DisableAutoLayer(appState *state.State) {
	detectArea := appState.GetDetectArea()
	if detectArea != nil {
		detectArea.Destroy()
		appState.SetDetectArea(nil)
	}

	window := appState.GetWindow()

	appState.RemoveSignalHandler("enter", window)
	appState.RemoveSignalHandler("leave", window)
}
