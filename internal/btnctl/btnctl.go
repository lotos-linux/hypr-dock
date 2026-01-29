package btnctl

import (
	defaultcontrol "hypr-dock/internal/defaultControl"
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func Dispatch(item *item.Item, appState *state.State) {
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

	// clickes
	ctrl.ResetSingle(func() {
		client, ok := utils.GetSingleValue(item.Windows)
		if ok {
			ipc.Hyprctl("dispatch focuswindow address:" + client.Address)

			showTimer.Stop()
			if pv.GetActive() {
				pv.Hide()
			}
		}
	})

	ctrl.ResetMulti(func() {
		if !pv.GetActive() {
			pv.Show(item, settings)
			pv.SetCurrentClass(item.ClassName)
		}
	})

	ctrl.OnContextOpen(func() {
		showTimer.Stop()
		if pv.GetActive() {
			pv.Hide()
		}
	})

	ctrl.Init()

	// hover
	item.Button.Connect("enter-notify-event", func() {
		instances := len(item.Windows)
		if instances == 0 {
			return
		}

		hideTimer.Stop()

		if pv.GetActive() && pv.HasClassChanged(item.ClassName) {
			moveTimer.Stop()
			moveTimer.Run(settings.PreviewAdvanced.MoveDelay, func() { pv.Change(item, settings) })
			pv.SetCurrentClass(item.ClassName)
			return
		}

		if !pv.GetActive() {
			showTimer.Run(settings.PreviewAdvanced.ShowDelay, func() { pv.Show(item, settings) })
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
			hideTimer.Run(settings.PreviewAdvanced.HideDelay, pv.Hide)
		}
	})

	// send to host control signal for auto mode
	pv.OnEnter(func(w *gtk.Window, e *gdk.Event) {
		appState.GetLayerctl().SendFocus()
	})

	pv.OnLeave(func(w *gtk.Window, e *gdk.Event) {
		appState.GetLayerctl().SendUnfocus()
	})

	pv.OnEmpty(func() {
		appState.GetLayerctl().SendUnfocus()
	})
}
