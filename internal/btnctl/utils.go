package btnctl

import (
	"hypr-dock/internal/item"
	"hypr-dock/internal/layering"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func connectContextMenu(item *item.Item, appState *state.State) {
	settings := appState.GetSettings()

	item.Button.Connect("button-release-event", func(button *gtk.Button, e *gdk.Event) {
		event := gdk.EventButtonNewFromEvent(e)
		if event.Button() == 3 {
			menu, err := item.ContextMenu(settings)
			if err != nil {
				log.Println(err)
				return
			}

			win, zone, err := getActivateZone(item.Button, settings.ContextPos, settings.Position)
			if err != nil {
				log.Println(err)
				return
			}

			firstg, secondg := getGravity(settings.Position)
			menu.PopupAtRect(win, zone, firstg, secondg, nil)
			ipc.DispatchEvent("hd>>open-context")
			menu.Connect("deactivate", func() {
				ipc.DispatchEvent("hd>>close-context")
				dispather(appState, item.Button)
			})

			return
		}
	})
}

func leftClick(btn *gtk.Button, handler func(e *gdk.Event)) {
	btn.Connect("button-release-event", func(button *gtk.Button, e *gdk.Event) {
		event := gdk.EventButtonNewFromEvent(e)
		if event.Button() != 3 {
			handler(e)
		}
	})
}

func dispather(appState *state.State, btn *gtk.Button) {
	window := appState.GetWindow()
	btn.SetStateFlags(gtk.STATE_FLAG_NORMAL, true)
	if appState.GetSettings().Layer == "auto" {
		layering.DispathLeaveEvent(window, nil, appState)
	}
}

func getActivateZone(v *gtk.Button, margin int, pos string) (*gdk.Window, *gdk.Rectangle, error) {
	var rect *gdk.Rectangle

	win, err := v.GetWindow()
	if err != nil {
		return nil, nil, err
	}

	switch pos {
	case "bottom":
		rect = gdk.RectangleNew(
			v.GetAllocation().GetX(),
			0-margin,
			v.GetAllocatedWidth(),
			v.GetAllocatedHeight(),
		)
	case "left":
		rect = gdk.RectangleNew(
			0-(v.GetAllocatedWidth()/2)-v.GetAllocation().GetX()+win.WindowGetWidth()+margin,
			v.GetAllocation().GetY(),
			v.GetAllocatedWidth(),
			v.GetAllocatedHeight(),
		)
	case "top":
		rect = gdk.RectangleNew(
			v.GetAllocation().GetX(),
			0-(v.GetAllocatedHeight()/2)-v.GetAllocation().GetY()+win.WindowGetHeight()+margin,
			v.GetAllocatedWidth(),
			v.GetAllocatedHeight(),
		)
	case "right":
		rect = gdk.RectangleNew(
			0-margin,
			v.GetAllocation().GetY(),
			v.GetAllocatedWidth(),
			v.GetAllocatedHeight(),
		)
	}

	return win, rect, err
}

func getGravity(pos string) (gdk.Gravity, gdk.Gravity) {
	var first, second gdk.Gravity

	switch pos {
	case "bottom":
		first = gdk.GDK_GRAVITY_NORTH
		second = gdk.GDK_GRAVITY_SOUTH
	case "left":
		first = gdk.GDK_GRAVITY_EAST
		second = gdk.GDK_GRAVITY_WEST
	case "top":
		second = gdk.GDK_GRAVITY_NORTH
		first = gdk.GDK_GRAVITY_SOUTH
	case "right":
		second = gdk.GDK_GRAVITY_EAST
		first = gdk.GDK_GRAVITY_WEST
	}

	return first, second
}
