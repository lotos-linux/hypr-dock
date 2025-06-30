package btnctl

import (
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
	"log"

	"github.com/gotk3/gotk3/gdk"
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
	pvState := appState.GetPVState()
	pv := pvState.GetPV()
	showTimer := pvState.GetShowTimer()
	hideTimer := pvState.GetHideTimer()
	moveTimer := pvState.GetMoveTimer()

	show := func() {
		pv.Show(item, settings)
		pvState.SetActive(true)
	}

	hide := func() {
		pv.Hide(item, settings)
		pvState.SetActive(false)
	}

	move := func() {
		pv.Move(item, settings)
	}

	leftClick(item.Button, func(e *gdk.Event) {
		if item.Instances == 0 {
			utils.Launch(item.DesktopData.Exec)
		}
		if item.Instances == 1 {
			ipc.Hyprctl("dispatch focuswindow address:" + item.Windows[0]["Address"])
		}
		if item.Instances > 1 {
			showTimer.Run(0, show)
			pvState.SetCurrentClass(item.ClassName)
		}
	})

	item.Button.Connect("enter-notify-event", func() {
		if item.Instances == 0 {
			return
		}

		hideTimer.Stop()

		if pvState.GetActive() && pvState.HasClassChanged(item.ClassName) {
			moveTimer.Stop()
			moveTimer.Run(settings.PreviewAdvanced.MoveDelay, move)
			pvState.SetCurrentClass(item.ClassName)
			return
		}

		if !pvState.GetActive() {
			showTimer.Run(settings.PreviewAdvanced.ShowDelay, show)
			pvState.SetCurrentClass(item.ClassName)
		}
	})

	item.Button.Connect("leave-notify-event", func() {
		if item.Instances == 0 {
			return
		}

		showTimer.Stop()
		if pvState.GetActive() {
			hideTimer.Run(settings.PreviewAdvanced.HideDelay, hide)
		}
	})
}

func defaultControl(item *item.Item, appState *state.State) {
	settings := appState.GetSettings()

	leftClick(item.Button, func(e *gdk.Event) {
		if item.Instances == 0 {
			utils.Launch(item.DesktopData.Exec)
		}
		if item.Instances == 1 {
			ipc.Hyprctl("dispatch focuswindow address:" + item.Windows[0]["Address"])
		}
		if item.Instances > 1 {
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
