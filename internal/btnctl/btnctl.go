package btnctl

import (
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

func Dispatch(item *item.Item, appState *state.State) {
	connectContextMenu(item, appState)

	if appState.GetSettings().Preview == "none" {
		defaultControl(item, appState)
		return
	}

	previewControl(item, appState)
}

func previewControl(item *item.Item, appState *state.State) {
	settings := appState.GetSettings()
	pv := appState.GetPV()
	showTimer := pv.GetShowTimer()
	hideTimer := pv.GetHideTimer()
	moveTimer := pv.GetMoveTimer()

	show := func() {
		glib.IdleAdd(func() {
			pv.Show(item, settings)
		})
		pv.SetActive(true)
	}

	hide := func() {
		glib.IdleAdd(func() {
			pv.Hide()
		})
		pv.SetActive(false)
	}

	move := func() {
		glib.IdleAdd(func() {
			pv.Change(item, settings)
		})
	}

	ipc.AddEventListener("hd>>open-context", func(e string) {
		showTimer.Stop()
		if pv.GetActive() {
			hideTimer.Run(0, hide)
		}
	}, true)

	leftClick(item.Button, func(e *gdk.Event) {
		instances := len(item.Windows)

		if instances == 0 {
			item.App.Run()
		}
		if instances == 1 {
			client, ok := utils.GetSingleValue(item.Windows)
			if ok {
				ipc.Hyprctl("dispatch focuswindow address:" + client.Address)
				ipc.DispatchEvent("hd>>focus-window")
			}
		}
		if instances > 1 {
			if !pv.GetActive() {
				showTimer.Run(0, show)
				pv.SetCurrentClass(item.ClassName)
			}
		}
	})

	item.Button.Connect("enter-notify-event", func() {
		instances := len(item.Windows)
		if instances == 0 {
			return
		}

		hideTimer.Stop()

		if pv.GetActive() && pv.HasClassChanged(item.ClassName) {
			// fmt.Println("if true")
			moveTimer.Stop()
			moveTimer.Run(settings.PreviewAdvanced.MoveDelay, move)
			pv.SetCurrentClass(item.ClassName)
			return
		}

		if !pv.GetActive() {
			showTimer.Run(settings.PreviewAdvanced.ShowDelay, show)
			pv.SetCurrentClass(item.ClassName)
		}
	})

	item.Button.Connect("leave-notify-event", func() {
		instances := len(item.Windows)
		if instances == 0 {
			return
		}

		showTimer.Stop()
		if pv.GetActive() {
			hideTimer.Run(settings.PreviewAdvanced.HideDelay, hide)
		}
	})
}

func defaultControl(item *item.Item, appState *state.State) {
	settings := appState.GetSettings()

	leftClick(item.Button, func(e *gdk.Event) {
		instances := len(item.Windows)

		if instances == 0 {
			item.App.Run()
		}
		if instances == 1 {
			client, ok := utils.GetSingleValue(item.Windows)
			if ok {
				ipc.Hyprctl("dispatch focuswindow address:" + client.Address)
			}
		}
		if instances > 1 {
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
	})
}
