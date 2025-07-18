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

func New(item *item.Item, settings settings.Settings, onReady func(w, h int)) (box *gtk.Box, err error) {
	spacing := settings.PreviewStyle.Spacing

	wrapper, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, settings.ContextPos)
	if err != nil {
		return nil, err
	}
	wrapper.SetName("pv-wrap")

	var (
		totalWidth   int
		readyCount   int
		commonHeight int
		mutex        sync.Mutex
	)

	for _, window := range item.Windows {
		windowBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		if err != nil {
			log.Println(err)
			continue
		}

		windowBox.SetName("pv-item")

		eventBox, err := gtk.EventBoxNew()
		if err != nil {
			log.Println(err)
			continue
		}

		eventBox.SetName("pv-event-box")

		windowBoxContent, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		if err != nil {
			log.Println(err)
			continue
		}

		windowBoxContent.SetMarginBottom(spacing)
		windowBoxContent.SetMarginEnd(spacing)
		windowBoxContent.SetMarginStart(spacing)
		windowBoxContent.SetMarginTop(spacing)

		titleBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
		if err != nil {
			log.Println(err)
			continue
		}

		titleBox.SetMarginBottom(5)

		icon, err := utils.CreateImage(item.DesktopData.Icon, 16)
		if err != nil {
			log.Println(err)
			continue
		}

		label, err := gtk.LabelNew(window["Title"])
		if err != nil {
			log.Println(err)
			continue
		}

		label.SetEllipsize(pango.ELLIPSIZE_END)
		label.SetXAlign(0)
		label.SetHExpand(true)
		label.SetMarginBottom(2)

		closeBtn, err := gtk.ButtonNewFromIconName("close", gtk.ICON_SIZE_SMALL_TOOLBAR)
		if err != nil {
			log.Println(err)
			continue
		}
		closeBtn.SetName("close-btn")
		utils.AddStyle(closeBtn, "#close-btn {padding: 0;}")

		eventBox.Connect("button-press-event", func(eb *gtk.EventBox, e *gdk.Event) {
			go ipc.Hyprctl("dispatch focuswindow address:" + window["Address"])
			go ipc.DispatchEvent("hd>>focus-window")
		})

		context, err := windowBox.GetStyleContext()
		if err == nil {
			eventBox.Connect("enter-notify-event", func() {
				context.AddClass("hover")
			})
			eventBox.Connect("leave-notify-event", func(w *gtk.Window, e *gdk.Event) {
				event := gdk.EventCrossingNewFromEvent(e)
				isInWindow := event.Detail() == 3 || event.Detail() == 0

				if isInWindow {
					context.RemoveClass("hover")
				}
			})
		}

		display, err := gdk.DisplayGetDefault()
		if err == nil {
			pointer, _ := gdk.CursorNewFromName(display, "pointer")
			arrow, _ := gdk.CursorNewFromName(display, "default")

			eventBox.Connect("enter-notify-event", func() {
				win, _ := eventBox.GetWindow()
				if win != nil {
					win.SetCursor(pointer)
				}
			})

			eventBox.Connect("leave-notify-event", func(eb *gtk.EventBox, e *gdk.Event) {
				event := gdk.EventCrossingNewFromEvent(e)
				win, _ := eventBox.GetWindow()

				if win != nil && event.Detail() != 2 {
					win.SetCursor(arrow)
				}
			})
		}

		stream, err := hysc.StreamNew(window["Address"])
		if err != nil {
			log.Println(err)
			continue
		}

		stream.OnReady(func(s *hysc.Size) {
			if s == nil {
				return
			}

			closeBtn.Connect("button-press-event", func() {
				go ipc.Hyprctl("dispatch closewindow address:" + window["Address"])
				if item.Instances == 1 {
					go ipc.DispatchEvent("hd>>focus-window")
					return
				}

				newW := totalWidth - s.W - spacing*2 - settings.ContextPos

				go ipc.DispatchEvent(fmt.Sprintf("hd>>close-window>>%d::%d", newW, s.H))
				windowBox.Destroy()
				wrapper.ShowAll()
			})

			glib.IdleAdd(func() {
				mutex.Lock()
				defer mutex.Unlock()

				totalWidth += s.W
				readyCount++
				commonHeight = s.H

				if readyCount == len(item.Windows) {
					totalWidth = totalWidth + settings.ContextPos*(len(item.Windows)-1) + 2*spacing*len(item.Windows)
					commonHeight = commonHeight + 2*spacing
					onReady(totalWidth, commonHeight)
				}
			})
		})

		stream.SetHScale(settings.PreviewStyle.Size)
		stream.SetBorderRadius(settings.PreviewStyle.BorderRadius)

		if settings.Preview == "live" {
			err = stream.Start(settings.PreviewAdvanced.FPS, settings.PreviewAdvanced.BufferSize)
		} else {
			err = stream.CaptureFrame()
		}

		if err != nil {
			log.Println(err)
			continue
		}

		titleBox.Add(icon)
		titleBox.Add(label)
		titleBox.Add(closeBtn)

		windowBoxContent.Add(titleBox)
		windowBoxContent.Add(stream)

		eventBox.Add(windowBoxContent)
		windowBox.Add(eventBox)
		wrapper.Add(windowBox)
	}

	return wrapper, nil
}
