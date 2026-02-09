package app

import (
	"os"
	"slices"

	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/btnctl"
	"hypr-dock/internal/hypr/hyprOpt"
	"hypr-dock/internal/item"
	"hypr-dock/internal/state"
	"hypr-dock/pkg/ipc"
)

func BuildApp(appState *state.State) *gtk.Box {
	settings := appState.GetSettings()
	orientation := appState.GetLayerctl().GetOrientation()
	log := appState.GetLogger()

	app, err := gtk.BoxNew(orientation, 0)
	if err != nil {
		log.Error("Unable to create gtk box:", "package", "app", "err", err)
		os.Exit(2)
	}

	initMargin(app, appState)
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

	pin := slices.Contains(*appState.GetPinned(), className)
	added := list.Get(className) != nil

	if !pin && !added {
		InitNewItemInClass(className, appState)
	}

	list.Get(className).AddWindow(ipcClient)
	appState.GetWindow().ShowAll()
}

func InitNewItemInClass(className string, appState *state.State) {
	log := appState.GetLogger()

	list := appState.GetList()
	item, err := item.New(className, appState.GetSettings())
	if err != nil {
		log.Error("Unable to creat app item", "err", err)
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
	item, _, err := appState.GetList().SearchWindow(address)
	if err != nil {
		return
	}

	lastWindow := len(item.Windows) == 1
	pin := slices.Contains(*appState.GetPinned(), item.ClassName)

	if lastWindow && !pin {
		item.Remove()
		return
	}

	item.RemoveWindow(address)

	appState.GetWindow().ShowAll()
}

func ChangeWindowTitle(address string, title string, appState *state.State) {
	_, client, err := appState.GetList().SearchWindow(address)
	if err != nil {
		return
	}

	client.Title = title
}

func initMargin(app *gtk.Box, appState *state.State) {
	log := appState.GetLogger()

	settings := appState.GetSettings()
	position := settings.Position
	defMargin := settings.Margin

	if !settings.SystemGapUsed {
		setMargin(app, position, defMargin)
		return
	}

	margin, err := hyprOpt.GetGap()
	if err != nil {
		log.Error("Failed to get gaps, indent set from settings", "error", err)
		setMargin(app, position, defMargin)
	}

	setMargin(app, position, margin...)

	hyprOpt.GapChangeEvent(func(gaps []int) {
		setMargin(app, position, gaps...)
	})
}

func setMargin(app *gtk.Box, position string, margin ...int) {
	if len(margin) == 1 {
		switch position {
		case "bottom":
			app.SetMarginBottom(margin[0])
		case "left":
			app.SetMarginStart(margin[0])
		case "right":
			app.SetMarginEnd(margin[0])
		case "top":
			app.SetMarginTop(margin[0])
		}
	}

	if len(margin) == 4 {
		switch position {
		case "bottom":
			app.SetMarginBottom(margin[0])
		case "left":
			app.SetMarginStart(margin[1])
		case "right":
			app.SetMarginEnd(margin[2])
		case "top":
			app.SetMarginTop(margin[3])
		}
	}
}

// func addWindowMarginRule(app *gtk.Box, appState *state.State) {
// 	settings := appState.GetSettings()
// 	position := settings.Position
// 	var marginProvider *gtk.CssProvider

// 	switch settings.SystemGapUsed {
// 	case true:
// 		margin, err := hyprOpt.GetGap()
// 		if err != nil {
// 			log.Println(err, "\nSet margin in config")
// 			applyWindowMarginCSS(app, position, settings.Margin)
// 		}

// 		marginProvider = applyWindowMarginCSS(app, position, margin[0])

// 		hyprOpt.GapChangeEvent(func(gap int) {
// 			utils.RemoveStyleProvider(app, marginProvider)
// 			marginProvider = applyWindowMarginCSS(app, position, gap)
// 			log.Println("Window margins updated successfully: ", gap)
// 		})
// 	case false:
// 		applyWindowMarginCSS(app, position, settings.Margin)
// 	}
// }

// func applyWindowMarginCSS(app *gtk.Box, position string, margin int) *gtk.CssProvider {
// 	css := fmt.Sprintf("#app {margin-%s: %dpx;}", position, margin)

// 	marginProvider, err := gtk.CssProviderNew()
// 	if err != nil {
// 		log.Printf("Failed to create CSS provider: %v", err)
// 		return nil
// 	}

// 	appStyleContext, err := app.GetStyleContext()
// 	if err != nil {
// 		log.Printf("Failed to get style context: %v", err)
// 		return nil
// 	}

// 	appStyleContext.AddProvider(marginProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

// 	err = marginProvider.LoadFromData(css)
// 	if err != nil {
// 		log.Printf("Failed to load CSS data: %v", err)
// 		return nil
// 	}

// 	return marginProvider
// }
