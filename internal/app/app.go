package app

import (
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/btnctl"
	"hypr-dock/internal/hypr/hyprOpt"
	"hypr-dock/internal/item"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
)

func BuildApp(appState *state.State) *gtk.Box {
	settings := appState.GetSettings()
	orientation := appState.GetOrientation()

	app, err := gtk.BoxNew(orientation, 0)
	if err != nil {
		log.Println("BuildApp() | app | gtk.BoxNew()")
		log.Fatal(err)
	}

	addWindowMarginRule(app, appState)
	app.SetName("app")

	itemsBox, _ := gtk.BoxNew(orientation, settings.Spacing)
	itemsBox.SetName("items-box")

	switch orientation {
	case gtk.ORIENTATION_HORIZONTAL:
		itemsBox.SetMarginEnd(int(float64(settings.Spacing) * 0.8))
		itemsBox.SetMarginStart(int(float64(settings.Spacing) * 0.8))
	case gtk.ORIENTATION_VERTICAL:
		itemsBox.SetMarginBottom(int(float64(settings.Spacing) * 0.8))
		itemsBox.SetMarginTop(int(float64(settings.Spacing) * 0.8))
	}

	appState.SetItemsBox(itemsBox)
	renderItems(appState)
	app.Add(itemsBox)

	return app
}

func renderItems(appState *state.State) {
	clients, _ := ipc.GetClients()

	for _, className := range *appState.GetPinned() {
		InitNewItemInClass(className, appState)
	}

	for _, ipcClient := range clients {
		InitNewItemInIPC(ipcClient, appState)
	}

	ipc.DispatchEvent("hd>>dock-render-finish")
}

func InitNewItemInIPC(ipcClient ipc.Client, appState *state.State) {
	list := appState.GetList()
	className := ipcClient.Class
	if !slices.Contains(*appState.GetPinned(), className) && list.Get(className) == nil {
		InitNewItemInClass(className, appState)
	}

	list.Get(className).UpdateState(ipcClient, appState.GetSettings())
	appState.GetWindow().ShowAll()
}

func InitNewItemInClass(className string, appState *state.State) {
	list := appState.GetList()
	item, err := item.New(className, appState.GetSettings())
	if err != nil {
		log.Println(err)
		return
	}

	btnctl.Dispatch(item, appState)

	item.List = list.GetMap()
	item.PinnedList = appState.GetPinned()
	list.Add(className, item)

	appState.GetItemsBox().Add(item.ButtonBox)
	appState.GetWindow().ShowAll()
}

func RemoveApp(address string, appState *state.State) {
	item, windowIndex, err := searchByAddress(address, appState)
	if err != nil {
		log.Println(err)
		return
	}

	className := item.ClassName
	if item.Instances == 1 && !slices.Contains(*appState.GetPinned(), className) {
		item.Remove()
		return
	}

	item.RemoveLastInstance(windowIndex, appState.GetSettings())

	appState.GetWindow().ShowAll()
}

func searchByAddress(address string, appState *state.State) (*item.Item, int, error) {
	for _, item := range appState.GetList().GetMap() {
		for windowIndex, window := range item.Windows {
			if window["Address"] == address {
				return item, windowIndex, nil
			}
		}
	}

	err := errors.New("Window not found: " + address)
	return nil, 0, err
}

func ChangeWindowTitle(address string, title string, appState *state.State) {
	for _, data := range appState.GetList().GetMap() {
		for _, appWindow := range data.Windows {
			if appWindow["Address"] == address {
				appWindow["Title"] = title
			}
		}
	}
}

func addWindowMarginRule(app *gtk.Box, appState *state.State) {
	settings := appState.GetSettings()
	position := settings.Position
	var marginProvider *gtk.CssProvider

	switch settings.SystemGapUsed {
	case "true":
		margin, err := hyprOpt.GetGap()
		if err != nil {
			log.Println(err, "\nSet margin in config")
			applyWindowMarginCSS(app, position, settings.Margin)
		}

		marginProvider = applyWindowMarginCSS(app, position, margin[0])

		hyprOpt.GapChangeEvent(func(gap int) {
			utils.RemoveStyleProvider(app, marginProvider)
			marginProvider = applyWindowMarginCSS(app, position, gap)
			log.Println("Window margins updated successfully: ", gap)
		})
	case "false":
		applyWindowMarginCSS(app, position, settings.Margin)
	}
}

func applyWindowMarginCSS(app *gtk.Box, position string, margin int) *gtk.CssProvider {
	css := fmt.Sprintf("#app {margin-%s: %dpx;}", position, margin)

	marginProvider, err := gtk.CssProviderNew()
	if err != nil {
		log.Printf("Failed to create CSS provider: %v", err)
		return nil
	}

	appStyleContext, err := app.GetStyleContext()
	if err != nil {
		log.Printf("Failed to get style context: %v", err)
		return nil
	}

	appStyleContext.AddProvider(marginProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	err = marginProvider.LoadFromData(css)
	if err != nil {
		log.Printf("Failed to load CSS data: %v", err)
		return nil
	}

	return marginProvider
}
