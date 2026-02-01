package layering

import (
	detectzone "hypr-dock/internal/detectZone"
	"hypr-dock/internal/pkg/timer"
	"hypr-dock/internal/settings"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Control struct {
	window   *gtk.Window
	settings *settings.Settings
	da       *detectzone.DetectArea

	smartEnter glib.SignalHandle
	smartLeave glib.SignalHandle

	hideTimer *timer.Timer
	special   bool

	orientation gtk.Orientation
	edge        layershell.LayerShellEdgeFlags

	layers map[string]layershell.LayerShellLayerFlags
}

func New(window *gtk.Window, settings *settings.Settings) *Control {
	layers := map[string]layershell.LayerShellLayerFlags{
		"background": layershell.LAYER_SHELL_LAYER_BACKGROUND,
		"bottom":     layershell.LAYER_SHELL_LAYER_BOTTOM,
		"top":        layershell.LAYER_SHELL_LAYER_TOP,
		"overlay":    layershell.LAYER_SHELL_LAYER_OVERLAY,
	}

	return &Control{
		window:   window,
		settings: settings,

		layers:    layers,
		hideTimer: timer.New(),
	}
}

func (c *Control) Init() {
	layershell.InitForWindow(c.window)
	layershell.SetNamespace(c.window, "hypr-dock")

	c.SetPosition()
	c.SetLayer()
}

func NewInit(window *gtk.Window, settings *settings.Settings) *Control {
	ctrl := New(window, settings)
	ctrl.Init()
	return ctrl
}

func (c *Control) SetLayer() {
	c.clear()

	if c.settings.SmartView {
		c.smart()
		return
	}

	if c.settings.Exclusive {
		layershell.AutoExclusiveZoneEnable(c.window)
	}

	layershell.SetLayer(c.window, c.layers[c.settings.Layer])
}

func (c *Control) SetPosition() {
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

	position := c.settings.Position

	layershell.SetAnchor(c.window, edges[position], true)
	layershell.SetMargin(c.window, edges[position], 0)

	c.orientation = oreintations[position]
	c.edge = edges[position]
}

func (c *Control) smart() {
	layershell.SetLayer(c.window, c.layers["bottom"])

	c.da = detectzone.New(c.window, c.settings)
	c.da.OnEnter(func() {
		c.SendFocus()
	})
	c.da.OnLeave(func() {
		c.SendUnfocus()
	})

	c.smartEnter = c.window.Connect("enter-notify-event", func(_ *gtk.Window, e *gdk.Event) {
		if !is_e3e4(e) || c.special {
			return
		}

		c.SendFocus()
	})

	c.smartLeave = c.window.Connect("leave-notify-event", func(_ *gtk.Window, e *gdk.Event) {
		if !is_e3e4(e) {
			return
		}

		c.SendUnfocus()
	})
}

func (c *Control) clear() {
	layershell.SetExclusiveZone(c.window, 0)

	if c.da != nil {
		c.da.Destroy()
		c.da = nil
	}

	if c.smartEnter > 0 {
		c.window.HandlerDisconnect(c.smartEnter)
	}

	if c.smartLeave > 0 {
		c.window.HandlerDisconnect(c.smartLeave)
	}
}

func (c *Control) SendUnfocus() {
	if !c.settings.SmartView {
		return
	}

	c.hideTimer.Run(c.settings.AutoHideDelay, func() {
		layershell.SetLayer(c.window, layershell.LAYER_SHELL_LAYER_BOTTOM)
	})
}

func (c *Control) SendFocus() {
	if !c.settings.SmartView {
		return
	}

	c.hideTimer.Stop()
	layershell.SetLayer(c.window, c.layers["top"])
}

func (c *Control) SetSpecial(is bool) {
	c.special = is
}

func (c *Control) GetSpecial() bool {
	return c.special
}

func (c *Control) GetOrientation() gtk.Orientation {
	return c.orientation
}

func (c *Control) GetEdge() layershell.LayerShellEdgeFlags {
	return c.edge
}

func is_e3e4(e *gdk.Event) bool {
	ec := gdk.EventCrossingNewFromEvent(e)
	return ec.Detail() == 3 || ec.Detail() == 4
}
