package btnctl

import (
	defaultcontrol "hypr-dock/internal/defaultControl"
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"

	"github.com/gotk3/gotk3/glib"
)

func Dispatch(item *item.Item, appState *state.State) {
	defaultcontrol.ConnectContextMenu(item, appState)
	ctrl := defaultcontrol.New(item, appState)

	if appState.GetSettings().Preview != "none" {
		previewControl(item, ctrl, appState)
		return
	}

	ctrl.Init()
}

func previewControl(item *item.Item, ctrl *defaultcontrol.Control, appState *state.State) {
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

	ctrl.ResetSingle(func() {
		client, ok := utils.GetSingleValue(item.Windows)
		if ok {
			ipc.Hyprctl("dispatch focuswindow address:" + client.Address)
			ipc.DispatchEvent("hd>>focus-window")
		}
	})

	ctrl.ResetMulti(func() {
		if !pv.GetActive() {
			showTimer.Run(0, show)
			pv.SetCurrentClass(item.ClassName)
		}
	})

	ctrl.Init()

	ipc.AddEventListener("hd>>open-context", func(e string) {
		showTimer.Stop()
		if pv.GetActive() {
			hideTimer.Run(0, hide)
		}
	}, true)

	item.Button.Connect("enter-notify-event", func() {
		instances := len(item.Windows)
		if instances == 0 {
			return
		}

		hideTimer.Stop()

		if pv.GetActive() && pv.HasClassChanged(item.ClassName) {
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
