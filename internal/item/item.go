package item

import (
	"log"
	"slices"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/desktop"
	layerinfo "hypr-dock/internal/layerInfo"
	"hypr-dock/internal/pkg/cfg"
	"hypr-dock/internal/pkg/indicator"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/settings"

	"hypr-dock/pkg/ipc"
)

type Item struct {
	Windows        map[string]*ipc.Client
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
		Windows:        map[string]*ipc.Client{},
		IndicatorImage: indicatorImage,
		Button:         button,
		ButtonBox:      item,
		App:            app,
		ClassName:      className,
		List:           nil,
		PinnedList:     nil,
	}, nil
}

func (item *Item) RemoveWindow(windowAddress string, settings settings.Settings) {
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

func (item *Item) AddWindow(ipcClient ipc.Client, settings settings.Settings) {
	if item.IndicatorImage != nil {
		item.IndicatorImage.Destroy()
	}

	item.Windows[ipcClient.Address] = &ipcClient
	instances := len(item.Windows)

	indicatorImage, err := indicator.New(instances, settings)
	if err == nil {
		appendInducator(item.ButtonBox, indicatorImage, settings.Position)
	}

	item.IndicatorImage = indicatorImage

	if instances != 0 && settings.Preview != "none" {
		item.Button.SetTooltipText("")
	}
}

func (item *Item) IsPinned() bool {
	return slices.Contains(*item.PinnedList, item.ClassName)
}

func (item *Item) TogglePin(settings settings.Settings) {
	list := item.PinnedList
	className := item.ClassName

	pin := item.IsPinned()
	running := len(item.Windows) > 0

	if pin {
		utils.RemoveFromSliceByValue(list, className)
		log.Println("Remove:", className)
	}

	if pin && !running {
		item.Remove()
	}

	if !pin {
		utils.AddToSlice(list, className)
		log.Println("Add:", className)
	}

	file := settings.PinnedPath
	err := cfg.ChangeJsonPinnedApps(*list, file)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Printf("File %s saved successfully! (%s)", file, className)
}

func (item *Item) Remove() {
	item.ButtonBox.Destroy()
	delete(item.List, item.ClassName)
}

type Position struct {
	X, Y       int
	CX, CY     int
	RelX, RelY int

	W, H    int
	Monitor *gdk.Monitor
}

func (Item *Item) GetCord(settings settings.Settings) (*Position, error) {
	margin := settings.ContextPos
	pos := settings.Position
	v := Item.Button

	result := &Position{
		RelX: v.GetAllocation().GetX(),
		RelY: v.GetAllocation().GetY(),

		W: v.GetAllocatedWidth(),
		H: v.GetAllocatedHeight(),

		X: 0,
		Y: 0,
	}

	// get main layer info
	dock, err := layerinfo.GetDock()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Add monitor offset
	monitors, err := ipc.GetMonitors()
	if err != nil {
		log.Println("Error getting monitors:", err)
	} else {
		for _, m := range monitors {
			if m.Name == dock.Monitor {
				result.X = m.X
				result.Y = m.Y
				break
			}
		}
	}

	// get coord with centring
	switch pos {
	case "bottom", "top":
		result.CX = result.X + dock.X + result.RelX + result.W/2
		result.CY = result.Y + margin + dock.H
	case "left", "right":
		result.CX = result.X + margin + dock.W
		result.CY = result.Y + dock.Y + result.RelY + result.H/2
	}

	log.Println(result.CY)

	// get absolute coord
	result.X = result.RelX + dock.X
	result.Y = result.RelY + dock.Y

	// Monitor
	display, err := gdk.DisplayGetDefault()
	if err != nil {
		return result, err
	}

	monitor, err := display.GetMonitorAtPoint(result.X, result.Y)
	if err != nil {
		return result, err
	}

	result.Monitor = monitor

	return result, nil
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
