package detectzone

import (
	"hypr-dock/internal/settings"
	"log"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type DetectArea struct {
	onEnter func()
	onLeave func()

	*gtk.Window
}

func New(mainWindow *gtk.Window, settings settings.Settings) *DetectArea {
	detectWindow, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("InitDetectArea(), gtk.WindowNew() | ", err)
	}

	da := &DetectArea{
		Window: detectWindow,
		onEnter: func() {
			log.Printf("Detect enter")
		},
		onLeave: func() {
			log.Printf("Detect leave")
		},
	}

	da.SetName("detect")
	da.SetSizeRequest(-1, 1)

	layershell.InitForWindow(da.Window)
	layershell.SetNamespace(da.Window, "dock-detect")
	layershell.SetLayer(da.Window, layershell.LAYER_SHELL_LAYER_TOP)

	da.clickes()

	selectEdges(da.Window, settings)

	da.ShowAll()

	return da
}

func (da *DetectArea) clickes() {
	da.Connect("enter-notify-event", func(detectWindow *gtk.Window, e *gdk.Event) {
		da.onEnter()
	})

	da.Connect("leave-notify-event", func(detectWindow *gtk.Window, e *gdk.Event) {
		da.onLeave()
	})
}

func (da *DetectArea) OnEnter(handler func()) {
	da.onEnter = handler
}

func (da *DetectArea) OnLeave(handler func()) {
	da.onLeave = handler
}

func selectEdges(window *gtk.Window, settings settings.Settings) {
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
