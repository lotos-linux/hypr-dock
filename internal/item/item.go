package item

import (
	"log"
	"slices"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/desktop"
	layerinfo "hypr-dock/internal/layerInfo"

	// "hypr-dock/internal/pkg/cfg"
	"hypr-dock/internal/pkg/indicator"
	"hypr-dock/internal/pkg/pinned"
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

	Settings   *settings.Settings
	List       map[string]*Item
	PinnedList *[]string
}

func New(className string, settings *settings.Settings) (*Item, error) {
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

		Settings:   settings,
		List:       nil,
		PinnedList: nil,
	}, nil
}

func (i *Item) RemoveWindow(windowAddress string) {
	if i.IndicatorImage != nil {
		i.IndicatorImage.Destroy()
	}

	delete(i.Windows, windowAddress)
	instances := len(i.Windows)

	newImage, err := indicator.New(instances, i.Settings)
	if err == nil {
		appendInducator(i.ButtonBox, newImage, i.Settings.Position)
	}
	i.IndicatorImage = newImage

	if instances == 0 && i.Settings.Preview.Mode != "none" {
		i.Button.SetTooltipText(i.App.GetName())
	}
}

func (i *Item) AddWindow(ipcClient ipc.Client) {
	if i.IndicatorImage != nil {
		i.IndicatorImage.Destroy()
	}

	i.Windows[ipcClient.Address] = &ipcClient
	instances := len(i.Windows)

	indicatorImage, err := indicator.New(instances, i.Settings)
	if err == nil {
		appendInducator(i.ButtonBox, indicatorImage, i.Settings.Position)
	}

	i.IndicatorImage = indicatorImage

	if instances != 0 && i.Settings.Preview.Mode != "none" {
		i.Button.SetTooltipText("")
	}
}

func (i *Item) IsPinned() bool {
	return slices.Contains(*i.PinnedList, i.ClassName)
}

func (i *Item) TogglePin() {
	list := i.PinnedList
	className := i.ClassName

	pin := i.IsPinned()
	running := len(i.Windows) > 0

	if pin {
		utils.RemoveFromSliceByValue(list, className)
		log.Println("Remove:", className)
	}

	if pin && !running {
		i.Remove()
	}

	if !pin {
		utils.AddToSlice(list, className)
		log.Println("Add:", className)
	}

	file := i.Settings.PinnedPath
	err := pinned.Save(file, *list)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Printf("File %s saved successfully! (%s)", file, className)
}

func (i *Item) Remove() {
	i.ButtonBox.Destroy()
	delete(i.List, i.ClassName)
}

type Position struct {
	X, Y       int
	CX, CY     int
	RelX, RelY int

	W, H    int
	Monitor *gdk.Monitor
}

func (i *Item) GetCord() (*Position, error) {
	margin := i.Settings.ContextPos
	pos := i.Settings.Position
	v := i.Button

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
