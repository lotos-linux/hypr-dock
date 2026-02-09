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

	if appState.GetSettings().Preview.Mode != "none" {
		previewControl(item, ctrl, appState)
		return
	}

	ctrl.Init()
}

func previewControl(item *item.Item, ctrl *defaultcontrol.Control, appState *state.State) {
	pv := appState.GetPV()
	showTimer := pv.GetShowTimer()

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
			pv.Show(item)
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
		pv.SmartOpen(item)
	})

	item.Button.Connect("leave-notify-event", func() {
		pv.SmartHide(item)
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
