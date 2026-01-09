package pvwidget

import (
	"fmt"
	"hypr-dock/internal/hysc"
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/settings"
	"hypr-dock/pkg/ipc"
	"log"
	"sync"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

type Widget struct {
	readyCount   int
	totalWidth   int
	commonHeight int
	mutex        sync.Mutex
	*gtk.Box
	settings settings.Settings
	item     *item.Item
	onReady  func(w, h int)
}

func New(item *item.Item, settings settings.Settings, onReady func(w, h int)) (*Widget, error) {
	wrapper, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, settings.ContextPos)
	if err != nil {
		return nil, err
	}
	wrapper.SetName("pv-wrap")

	widget := &Widget{
		Box:      wrapper,
		settings: settings,
		item:     item,
		onReady:  onReady,
	}

	for _, window := range item.Windows {
		err := widget.createWindowWidget(window)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	return widget, nil
}

func (w *Widget) createWindowWidget(window *ipc.Client) error {
	padding := w.settings.PreviewStyle.Padding

	windowBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return err
	}
	windowBox.SetName("pv-item")

	eventBox, err := gtk.EventBoxNew()
	if err != nil {
		return err
	}
	eventBox.SetName("pv-event-box")

	windowBoxContent, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return err
	}
	windowBoxContent.SetMarginBottom(padding)
	windowBoxContent.SetMarginEnd(padding)
	windowBoxContent.SetMarginStart(padding)
	windowBoxContent.SetMarginTop(padding / 2)

	titleBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	if err != nil {
		return err
	}
	titleBox.SetMarginBottom(padding / 2)

	icon, err := utils.CreateImage(w.item.App.GetIcon(), 16)
	if err != nil {
		return err
	}

	label, err := gtk.LabelNew(window.Title)
	if err != nil {
		return err
	}
	label.SetEllipsize(pango.ELLIPSIZE_END)
	label.SetXAlign(0)
	label.SetHExpand(true)
	label.SetTooltipText(window.Title)

	iconName := utils.GetFirstAvailableImage([]string{
		"close",
		"close-symbolic",
		"window-close",
		"window-close-symbolic",
	})

	closeBtn, err := gtk.ButtonNewFromIconName(iconName, gtk.ICON_SIZE_SMALL_TOOLBAR)
	if err != nil {
		return err
	}
	closeBtn.SetName("close-btn")
	utils.AddStyle(closeBtn, "#close-btn {padding: 0;}")

	eventBox.Connect("button-press-event", func(eb *gtk.EventBox, e *gdk.Event) {
		go ipc.Hyprctl("dispatch focuswindow address:" + window.Address)
		go ipc.DispatchEvent("hd>>focus-window")
	})

	context, err := windowBox.GetStyleContext()
	if err == nil {
		utils.SetAutoHover(eventBox.ToWidget(), context)
	}
	utils.SetCursorPointer(eventBox.ToWidget())

	stream, err := hysc.StreamNew(window.Address)
	if err != nil {
		return err
	}

	stream.OnReady(func(s *hysc.Size) {
		if s == nil {
			return
		}

		closeBtn.Connect("button-press-event", func() {
			go ipc.Hyprctl("dispatch closewindow address:" + window.Address)
			if len(w.item.Windows) == 1 {
				go ipc.DispatchEvent("hd>>focus-window")
				return
			}

			w.mutex.Lock()
			defer w.mutex.Unlock()

			w.totalWidth = w.totalWidth - s.W - padding*2 - w.settings.ContextPos

			go ipc.DispatchEvent(fmt.Sprintf("hd>>close-window>>%d", w.totalWidth))
			windowBox.Destroy()
			w.ShowAll()
		})

		glib.IdleAdd(func() {
			w.mutex.Lock()
			defer w.mutex.Unlock()

			w.totalWidth += s.W
			w.readyCount++
			w.commonHeight = s.H

			if w.readyCount == len(w.item.Windows) {
				w.totalWidth = w.totalWidth + w.settings.ContextPos*(len(w.item.Windows)-1) + 2*padding*len(w.item.Windows)
				w.commonHeight = w.commonHeight + 2*padding + 20

				w.onReady(w.totalWidth, w.commonHeight)
			}
		})
	})

	stream.SetHScale(w.settings.PreviewStyle.Size)
	stream.SetBorderRadius(w.settings.PreviewStyle.BorderRadius)

	if w.settings.Preview == "live" {
		err = stream.Start(w.settings.PreviewAdvanced.FPS, w.settings.PreviewAdvanced.BufferSize)
	} else {
		err = stream.CaptureFrame()
	}

	if err != nil {
		return err
	}

	titleBox.Add(icon)
	titleBox.Add(label)
	titleBox.Add(closeBtn)

	windowBoxContent.Add(titleBox)
	windowBoxContent.Add(stream)

	eventBox.Add(windowBoxContent)
	windowBox.Add(eventBox)
	w.Add(windowBox)

	return nil
}
