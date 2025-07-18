package pvctl

import (
	"fmt"
	"hypr-dock/internal/item"
	layerinfo "hypr-dock/internal/layerInfo"
	"hypr-dock/internal/pkg/popup"
	"hypr-dock/internal/pkg/timer"
	"hypr-dock/internal/pvwidget"
	"hypr-dock/internal/settings"
	"hypr-dock/pkg/ipc"
	"log"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type PV struct {
	active bool

	showTimer *timer.Timer
	hideTimer *timer.Timer
	moveTimer *timer.Timer

	className string
	popup     *popup.Popup
}

func New(settings settings.Settings) *PV {
	return &PV{
		className: "90348d332fvecs324csd4",

		showTimer: timer.New(),
		hideTimer: timer.New(),
		moveTimer: timer.New(),

		popup: popup.New(),
	}
}

func (pv *PV) Show(item *item.Item, settings settings.Settings) {
	if pv == nil || item == nil {
		fmt.Println("pv is nil:", pv == nil, "|", "item is nil:", item == nil)
		return
	}

	hide := func() {
		glib.IdleAdd(func() {
			pv.Hide()
		})
		pv.SetActive(false)
	}

	ipc.AddEventListener("hd>>focus-window", func(e string) {
		pv.showTimer.Stop()
		hide()
	}, true)

	ipc.AddEventListener("hd>>close-window", func(e string) {
		event := strings.TrimPrefix(e, "hd>>close-window>>")
		eventDetail := strings.SplitN(event, "::", 2)
		ws := eventDetail[0]
		// hs := eventDetail[1]
		w, _ := strconv.Atoi(ws)
		// h, _ := strconv.Atoi(hs)

		x, y, _ := getCord(item.Button, settings)

		pv.popup.Move(x-w/2, y)
	}, true)

	pv.popup.SetWinCallBack(func(w *gtk.Window) error {
		w.Connect("enter-notify-event", func() {
			pv.hideTimer.Stop()
		})
		w.Connect("leave-notify-event", func(w *gtk.Window, e *gdk.Event) {
			event := gdk.EventCrossingNewFromEvent(e)
			isInWindow := event.Detail() == 3 || event.Detail() == 4

			if !isInWindow {
				return
			}
			pv.hideTimer.Run(settings.PreviewAdvanced.HideDelay, hide)
		})
		return nil
	})

	widget, err := pvwidget.New(item, settings, func(w, h int) {
		setCord(w, h, item, settings, func(x, y int, startx, starty string) {
			pv.popup.Open(x, y, startx, starty)
		})
	})
	if err != nil {
		log.Println(err)
		return
	}

	pv.popup.Set(widget)
}

func (pv *PV) Hide() {
	pv.popup.Close()
}

func (pv *PV) Change(item *item.Item, settings settings.Settings) {
	if pv == nil || item == nil {
		fmt.Println("pv is nil:", pv == nil, "|", "item is nil:", item == nil)
		return
	}

	hide := func() {
		glib.IdleAdd(func() {
			pv.Hide()
		})
		pv.SetActive(false)
	}

	pv.popup.SetWinCallBack(func(w *gtk.Window) error {
		w.Connect("enter-notify-event", func() {
			pv.hideTimer.Stop()
		})
		w.Connect("leave-notify-event", func() {
			pv.hideTimer.Run(settings.PreviewAdvanced.HideDelay, hide)
		})
		return nil
	})

	widget, err := pvwidget.New(item, settings, func(w, h int) {
		setCord(w, h, item, settings, func(x, y int, startx, starty string) {
			pv.popup.Move(x, y)
		})
	})
	if err != nil {
		log.Println(err)
		return
	}

	pv.popup.Set(widget)
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
	pv.className = className
}

func getCord(v *gtk.Button, settings settings.Settings) (int, int, error) {
	margin := settings.ContextPos
	pos := settings.Position

	ex := strings.Contains(settings.Layer, "exclusive")

	dock, err := layerinfo.GetDock()
	if err != nil {
		log.Println(err)
	}

	var x int
	var y int

	switch pos {
	case "bottom":
		x = dock.X + v.GetAllocation().GetX() + v.GetAllocatedWidth()/2
		y = margin
	case "left":
		x = margin
		y = dock.Y + v.GetAllocation().GetY() + v.GetAllocatedHeight()/2
	case "top":
		x = dock.X + v.GetAllocation().GetX() + v.GetAllocatedWidth()/2
		y = margin
	case "right":
		x = margin
		y = dock.Y + v.GetAllocation().GetY() + v.GetAllocatedHeight()/2
	}

	if !ex {
		switch pos {
		case "bottom", "top":
			y = y + dock.H
		case "left", "right":
			x = x + dock.W
		}
	}

	return x, y, err
}

func setCord(w, h int, item *item.Item, settings settings.Settings, callBack func(x, y int, startx, starty string)) {
	x, y, _ := getCord(item.Button, settings)
	var startx, starty string

	switch settings.Position {
	case "bottom":
		startx = "left"
		starty = "bottom"
	case "right":
		startx = "right"
		starty = "top"
	case "left":
		startx = "left"
		starty = "top"
	case "top":
		startx = "left"
		starty = "top"
	}

	switch settings.Position {
	case "top", "bottom":
		x = x - w/2
	case "left", "right":
		y = y - h/2
	}

	callBack(x, y, startx, starty)
}
