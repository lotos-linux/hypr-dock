package defaultcontrol

import (
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hashicorp/go-hclog"
)

type Control struct {
	item     *item.Item
	appState *state.State
	log      hclog.Logger

	zeroHandler   func()
	singleHandler func()
	multiHandler  func()

	onContext func()
}

func New(item *item.Item, appState *state.State) *Control {
	settings := appState.GetSettings()
	log := appState.GetLogger()

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
			log.Error("Unable to create windows menu", "error", err)
			return
		}

		win, zone, err := getActivateZone(item.Button, settings.ContextPos, settings.Position)
		if err != nil {
			log.Error("Failed to get activate zone", "error", err)
			return
		}

		firstg, secondg := getGravity(settings.Position)
		menu.PopupAtRect(win, zone, firstg, secondg, nil)
		menu.Connect("deactivate", func() {
			item.Button.SetStateFlags(gtk.STATE_FLAG_NORMAL, true)
			appState.GetLayerctl().SendUnfocus()
		})
	}

	return &Control{
		item:     item,
		appState: appState,
		log:      log,

		zeroHandler:   zeroHandler,
		singleHandler: singleHandler,
		multiHandler:  multiHandler,
	}
}

func (c *Control) Init() {
	c.connectContextMenu()

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

func (c *Control) OnContextOpen(handler func()) {
	c.onContext = handler
}

func (c *Control) connectContextMenu() {
	appState := c.appState
	settings := appState.GetSettings()
	item := c.item

	item.Button.Connect("button-release-event", func(button *gtk.Button, e *gdk.Event) {
		event := gdk.EventButtonNewFromEvent(e)
		if event.Button() == 3 {
			menu, err := item.ContextMenu()
			if err != nil {
				c.log.Error("Unable to create context menu", "error", err)
				return
			}

			win, zone, err := getActivateZone(item.Button, settings.ContextPos, settings.Position)
			if err != nil {
				c.log.Error("Failed to get activate zone", "error", err)
				return
			}

			firstg, secondg := getGravity(settings.Position)
			menu.PopupAtRect(win, zone, firstg, secondg, nil)

			if c.onContext != nil {
				c.onContext()
			}

			menu.Connect("deactivate", func() {
				item.Button.SetStateFlags(gtk.STATE_FLAG_NORMAL, true)
				appState.GetLayerctl().SendUnfocus()
			})

			return
		}
	})
}
