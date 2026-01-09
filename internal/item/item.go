package item

import (
	"log"
	"slices"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/desktop"
	"hypr-dock/internal/pkg/cfg"
	"hypr-dock/internal/pkg/indicator"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/settings"

	"hypr-dock/pkg/ipc"
)

type Item struct {
	Windows        map[string]ipc.Client
	App            *desktop.App
	ClassName      string
	Button         *gtk.Button
	ButtonBox      *gtk.Box
	IndicatorImage *gtk.Image
	List           map[string]*Item
	PinnedList     *[]string
}

func New(className string, settings settings.Settings) (*Item, error) {
	app := desktop.New(className)

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
		image, err := utils.CreateImage(app.GetIcon(), settings.IconSize)
		if err == nil {
			button.SetImage(image)
		} else {
			log.Println(err)
		}

		button.SetName(className)

		button.SetTooltipText(app.GetName())

		utils.SetCursorPointer(button.ToWidget())

		item.Add(button)
	} else {
		log.Println(err)
	}

	return &Item{
		Windows:        make(map[string]ipc.Client),
		IndicatorImage: indicatorImage,
		Button:         button,
		ButtonBox:      item,
		App:            app,
		ClassName:      className,
		List:           nil,
		PinnedList:     nil,
	}, nil
}

func (item *Item) RemoveLastInstance(windowAddress string, settings settings.Settings) {
	if item.IndicatorImage != nil {
		item.IndicatorImage.Destroy()
	}

	delete(item.Windows, windowAddress)
	instances := len(item.Windows)

	newImage, err := indicator.New(instances, settings)
	if err == nil {
		appendInducator(item.ButtonBox, newImage, settings.Position)
	}
	item.IndicatorImage = newImage

	if instances == 0 && settings.Preview != "none" {
		item.Button.SetTooltipText(item.App.GetName())
	}
}

func (item *Item) UpdateState(ipcClient ipc.Client, settings settings.Settings) {
	if item.IndicatorImage != nil {
		item.IndicatorImage.Destroy()
	}

	item.Windows[ipcClient.Address] = ipcClient
	instances := len(item.Windows)

	indicatorImage, err := indicator.New(instances, settings)
	if err == nil {
		appendInducator(item.ButtonBox, indicatorImage, settings.Position)
	}

	item.IndicatorImage = indicatorImage

	log.Println(instances, settings.Preview)
	if instances != 0 && settings.Preview != "none" {
		item.Button.SetTooltipText("")
	}
}

func (item *Item) IsPinned() bool {
	return slices.Contains(*item.PinnedList, item.ClassName)
}

func (item *Item) TogglePin(settings settings.Settings) {

	if item.IsPinned() {
		utils.RemoveFromSliceByValue(item.PinnedList, item.ClassName)
		if len(item.Windows) == 0 {
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
