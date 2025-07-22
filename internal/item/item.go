package item

import (
	"log"
	"slices"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/pkg/cfg"
	"hypr-dock/internal/pkg/desktop"
	"hypr-dock/internal/pkg/indicator"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/settings"

	"hypr-dock/pkg/ipc"
)

type Item struct {
	Instances      int
	Windows        []map[string]string
	DesktopData    *desktop.Desktop
	Actions        []*desktop.Action
	ClassName      string
	Button         *gtk.Button
	ButtonBox      *gtk.Box
	IndicatorImage *gtk.Image
	List           map[string]*Item
	PinnedList     *[]string
}

func New(className string, settings settings.Settings) (*Item, error) {
	desktopData := desktop.New(className)

	orientation := gtk.ORIENTATION_VERTICAL
	switch settings.Position {
	case "left", "right":
		orientation = gtk.ORIENTATION_HORIZONTAL
	}

	item, err := gtk.BoxNew(orientation, 0)
	if err != nil {
		return nil, err
	}

	indicatorImage, err := indicator.New(0, settings)
	if err == nil {
		appendInducator(item, indicatorImage, settings.Position)
	} else {
		log.Println(err)
	}

	button, err := gtk.ButtonNew()
	if err == nil {
		image, err := utils.CreateImage(desktopData.Icon, settings.IconSize)
		if err == nil {
			button.SetImage(image)
		} else {
			log.Println(err)
		}

		button.SetName(className)
		// button.SetTooltipText(desktopData.Name)

		display, err := gdk.DisplayGetDefault()
		if err == nil {
			pointer, _ := gdk.CursorNewFromName(display, "pointer")
			arrow, _ := gdk.CursorNewFromName(display, "default")

			button.Connect("enter-notify-event", func() {
				win, _ := button.GetWindow()
				if win != nil {
					win.SetCursor(pointer)
				}
			})

			button.Connect("leave-notify-event", func() {
				win, _ := button.GetWindow()
				if win != nil {
					win.SetCursor(arrow)
				}
			})
		}

		item.Add(button)
	} else {
		log.Println(err)
	}

	actions, err := desktop.GetAppActions(className)
	if err != nil {
		log.Println(err)
	}

	return &Item{
		IndicatorImage: indicatorImage,
		Button:         button,
		ButtonBox:      item,
		DesktopData:    desktopData,
		Actions:        actions,
		Instances:      0,
		ClassName:      className,
		List:           nil,
		PinnedList:     nil,
	}, nil
}

func (item *Item) RemoveLastInstance(windowIndex int, settings settings.Settings) {
	if item.IndicatorImage != nil {
		item.IndicatorImage.Destroy()
	}

	newImage, err := indicator.New(item.Instances-1, settings)
	if err == nil {
		appendInducator(item.ButtonBox, newImage, settings.Position)
	}

	item.Instances -= 1
	item.Windows = utils.RemoveFromSlice(item.Windows, windowIndex)
	item.IndicatorImage = newImage
}

func (item *Item) UpdateState(ipcClient ipc.Client, settings settings.Settings) {
	appWindow := map[string]string{
		"Address": ipcClient.Address,
		"Title":   ipcClient.Title,
	}

	if item.IndicatorImage != nil {
		item.IndicatorImage.Destroy()
	}

	indicatorImage, err := indicator.New(item.Instances+1, settings)
	if err == nil {
		// item.ButtonBox.Add(indicatorImage)
		appendInducator(item.ButtonBox, indicatorImage, settings.Position)
	}

	item.Windows = append(item.Windows, appWindow)
	item.IndicatorImage = indicatorImage
	item.Instances += 1
}

func (item *Item) IsPinned() bool {
	return slices.Contains(*item.PinnedList, item.ClassName)
}

func (item *Item) TogglePin(settings settings.Settings) {

	if item.IsPinned() {
		utils.RemoveFromSliceByValue(item.PinnedList, item.ClassName)
		if item.Instances == 0 {
			item.ButtonBox.Destroy()
			delete(item.List, item.ClassName)
		}
		log.Println("Remove:", item.ClassName)
	} else {
		utils.AddToSlice(item.PinnedList, item.ClassName)
		log.Println("Add:", item.ClassName)
	}

	err := cfg.ChangeJsonPinnedApps(*item.PinnedList, settings.PinnedPath)
	if err != nil {
		log.Println("Error: ", err)
	} else {
		log.Println("File", settings.PinnedPath, "saved successfully!", item.ClassName)
	}
}

func (item *Item) Remove() {
	item.ButtonBox.Destroy()
	delete(item.List, item.ClassName)
}

func appendInducator(parent *gtk.Box, child *gtk.Image, pos string) {
	switch pos {
	case "left", "right":
		buf := child.GetPixbuf()
		newBuf, err := buf.RotateSimple(gdk.PIXBUF_ROTATE_COUNTERCLOCKWISE)
		if err != nil {
			return
		}
		child.SetFromPixbuf(newBuf)
	}

	switch pos {
	case "left", "top":
		parent.PackStart(child, false, false, 0)
		parent.ReorderChild(child, 0)
	case "bottom", "right":
		parent.PackEnd(child, false, false, 0)
	}
}
