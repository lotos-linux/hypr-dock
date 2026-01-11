package defaultcontrol

import (
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type Control struct {
	item     *item.Item
	appState *state.State

	zeroHandler   func()
	singleHandler func()
	multiHandler  func()
}

func New(item *item.Item, appState *state.State) *Control {
	settings := appState.GetSettings()

	zeroHandler := func() {
		item.App.Run()
	}

	singleHandler := func() {
		client, ok := utils.GetSingleValue(item.Windows)
		if ok {
			ipc.Hyprctl("dispatch focuswindow address:" + client.Address)
		}
	}

	multiHandler := func() {
		menu, err := item.WindowsMenu()
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
		menu.Connect("deactivate", func() {
			dispather(appState, item.Button)
		})
	}

	return &Control{
		item:     item,
		appState: appState,

		zeroHandler:   zeroHandler,
		singleHandler: singleHandler,
		multiHandler:  multiHandler,
	}
}

func (c *Control) Init() {
	leftClick(c.item.Button, func(e *gdk.Event) {
		instances := len(c.item.Windows)

		if instances == 0 {
			c.zeroHandler()
		}
		if instances == 1 {
			c.singleHandler()
		}
		if instances > 1 {
			c.multiHandler()
		}
	})
}

func (c *Control) ResetZero(newHandler func()) {
	c.zeroHandler = newHandler
}

func (c *Control) ResetSingle(newHandler func()) {
	c.singleHandler = newHandler
}

func (c *Control) ResetMulti(newHandler func()) {
	c.multiHandler = newHandler
}

func ConnectContextMenu(item *item.Item, appState *state.State) {
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
