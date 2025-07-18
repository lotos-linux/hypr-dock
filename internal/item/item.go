package item

import (
	"log"
	"slices"

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
	ClassName      string
	Button         *gtk.Button
	ButtonBox      *gtk.Box
	IndicatorImage *gtk.Image
	List           map[string]*Item
	PinnedList     *[]string
}

func New(className string, settings settings.Settings) (*Item, error) {
	desktopData := desktop.New(className)

	item, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, err
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
		button.SetTooltipText(desktopData.Name)

		item.Add(button)
	} else {
		log.Println(err)
	}

	indicatorImage, err := indicator.New(0, settings)
	if err == nil {
		item.Add(indicatorImage)
	} else {
		log.Println(err)
	}

	return &Item{
		IndicatorImage: indicatorImage,
		Button:         button,
		ButtonBox:      item,
		DesktopData:    desktopData,
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
		item.ButtonBox.Add(newImage)
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
		item.ButtonBox.Add(indicatorImage)
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
