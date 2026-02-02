package pvctl

import (
	"fmt"
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/popup"
	"hypr-dock/internal/pkg/timer"
	"hypr-dock/internal/pvwidget"
	"hypr-dock/internal/settings"
	"hypr-dock/pkg/ipc"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type PV struct {
	active bool

	showTimer *timer.Timer
	hideTimer *timer.Timer
	moveTimer *timer.Timer

	className    string
	preClassName string

	popup    *popup.Popup
	widget   *pvwidget.Widget
	settings *settings.Settings

	onEnter func(w *gtk.Window, e *gdk.Event)
	onLeave func(w *gtk.Window, e *gdk.Event)
	onEmpty func()
}

func New(settings *settings.Settings) *PV {
	return &PV{
		className:    "90348d332fvecs324csd4",
		preClassName: "",

		showTimer: timer.New(),
		hideTimer: timer.New(),
		moveTimer: timer.New(),

		onEmpty: func() { log.Printf("Debug: all window closed") },

		popup:    popup.New(),
		settings: settings,
	}
}

func (pv *PV) Show(item *item.Item) {
	glib.IdleAdd(func() {
		pv.show(item)
	})
	pv.SetActive(true)
}

func (pv *PV) Change(item *item.Item) {
	glib.IdleAdd(func() {
		pv.change(item)
	})
}

func (pv *PV) Hide() {
	glib.IdleAdd(func() {
		pv.hide()
	})
	pv.SetActive(false)
}

func (pv *PV) show(item *item.Item) {
	if pv == nil {
		fmt.Printf("Debug: pv is nil")
		return
	}

	if item == nil {
		fmt.Printf("Debug: item is nill")
		return
	}

	// widget settings
	widget, err := pvwidget.New(item, pv.settings)
	if err != nil {
		log.Println(err)
		return
	}

	widget.OnResize(func(w, h int) {
		pv.resize(item, w, h)
	})

	widget.OnClick(func(c *ipc.Client) {
		log.Printf("Debug: %s window focused (%s)", c.Title, c.Address)

		pv.showTimer.Stop()
		pv.Hide()
	})

	widget.OnEmpty(func() {
		pv.onEmpty()
		pv.showTimer.Stop()
		pv.Hide()
	})

	widget.OnReady(func(w, h int) {
		pv.popup.SetWinCallBack(func(window *gtk.Window) error {
			window.SetSizeRequest(-1, h+5)
			return pv.popupWinSet(window)
		})

		target, orig := prepareCord(w, h, item, pv.settings)
		if orig.Monitor != nil {
			pv.popup.SetMonitor(orig.Monitor)
		}

		err := pv.popup.Open(target.x, target.y, target.anchorX, target.anchorY)
		if err != nil {
			log.Println("Error: faild open preview popup:", err)
		}
	})

	pv.widget = widget
	pv.popup.Set(widget)
}

func (pv *PV) change(item *item.Item) {
	if pv.widget != nil && pv.widget.GetClass() == pv.className {
		return
	}

	if pv == nil {
		fmt.Printf("Debug: pv is nil")
		return
	}

	if item == nil {
		fmt.Printf("Debug: item is nill")
		return
	}

	// widget recreate
	widget, err := pvwidget.New(item, pv.settings)
	if err != nil {
		log.Println(err)
		return
	}

	widget.OnResize(func(w, h int) {
		pv.resize(item, w, h)
	})

	widget.OnClick(func(c *ipc.Client) {
		log.Printf("Debug: %s window focused (%s)", c.Title, c.Address)

		pv.showTimer.Stop()
		pv.Hide()
	})

	widget.OnEmpty(func() {
		pv.onEmpty()
		pv.showTimer.Stop()
		pv.Hide()
	})

	widget.OnReady(func(w, h int) {
		target, _ := prepareCord(w, h, item, pv.settings)
		pv.popup.Move(target.x, target.y)
	})

	pv.widget = widget
	pv.popup.Set(widget)
}

func (pv *PV) hide() {
	pv.popup.Close()
}

func (pv *PV) popupWinSet(w *gtk.Window) error {
	w.Connect("enter-notify-event", func(w *gtk.Window, e *gdk.Event) {
		pv.hideTimer.Stop()

		if pv.moveTimer.IsRunning() {
			pv.moveTimer.Stop()
			pv.className = pv.preClassName
		}

		if pv.onEnter != nil {
			pv.onEnter(w, e)
		}
	})
	w.Connect("leave-notify-event", func(w *gtk.Window, e *gdk.Event) {
		event := gdk.EventCrossingNewFromEvent(e)
		isInWindow := event.Detail() == 3 || event.Detail() == 4

		if !isInWindow {
			return
		}
		pv.hideTimer.Run(pv.settings.Preview.HideDelay, pv.Hide)

		if pv.onLeave != nil {
			pv.onLeave(w, e)
		}
	})
	return nil
}

func (pv *PV) resize(item *item.Item, w, h int) {
	horizontal := pv.settings.Position == "top" || pv.settings.Position == "bottom"

	if horizontal {
		// get item buttom cord
		cord, err := item.GetCord()
		if err != nil {
			log.Println(err)
		}

		// move
		pv.popup.Move(cord.CX-w/2, cord.CY)

		// hide if mouse not in popup (if in popup, enter pointer evets stoped timer)
		pv.hideTimer.Run(pv.settings.Preview.HideDelay, pv.Hide)
	}
}

func (pv *PV) OnEnter(handler func(w *gtk.Window, e *gdk.Event)) {
	pv.onEnter = handler
}

func (pv *PV) OnLeave(handler func(w *gtk.Window, e *gdk.Event)) {
	pv.onLeave = handler
}

func (pv *PV) OnEmpty(handler func()) {
	pv.onEmpty = handler
}

func (pv *PV) SetActive(flag bool) {
	pv.active = flag
}

func (pv *PV) GetActive() bool {
	return pv.active
}

func (pv *PV) GetShowTimer() *timer.Timer {
	return pv.showTimer
}

func (pv *PV) GetHideTimer() *timer.Timer {
	return pv.hideTimer
}

func (pv *PV) GetMoveTimer() *timer.Timer {
	return pv.moveTimer
}

func (pv *PV) HasClassChanged(className string) bool {
	return pv.className != className
}

func (pv *PV) SetCurrentClass(className string) {
	pv.preClassName = pv.className
	pv.className = className
}

type popupTarget struct {
	x, y             int
	anchorX, anchorY string
}

func prepareCord(w, h int, item *item.Item, settings *settings.Settings) (target popupTarget, orig *item.Position) {
	orig, err := item.GetCord()
	if err != nil {
		log.Println(err)
	}

	target = popupTarget{}

	// Anchor
	switch settings.Position {
	case "bottom":
		target.anchorX = "left"
		target.anchorY = "bottom"

	case "right":
		target.anchorX = "right"
		target.anchorY = "top"

	case "left", "top":
		target.anchorX = "left"
		target.anchorY = "top"
	}

	// Popup center
	switch settings.Position {
	case "bottom", "top":
		target.x = orig.CX - w/2
		target.y = orig.CY
	case "left", "right":
		target.y = orig.CY - h/2
		target.x = orig.CX
	}

	// Translate global (x, y) to relative (x - geo.X, y - geo.Y)
	if orig.Monitor != nil {
		geo := orig.Monitor.GetGeometry()
		target.x -= geo.GetX()
		target.y -= geo.GetY()
	}

	return target, orig
}
