package item

import (
	"errors"
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"

	"hypr-dock/internal/desktop"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/internal/settings"
	"hypr-dock/pkg/ipc"
)

func (item *Item) WindowsMenu() (*gtk.Menu, error) {
	menu, err := gtk.MenuNew()
	if err != nil {
		log.Println(err)
	}

	desktopData := desktop.New(item.ClassName)

	AddWindowsItemToMenu(menu, item.Windows, desktopData)

	menu.SetName("windows-menu")
	menu.ShowAll()

	return menu, nil
}

func (item *Item) ContextMenu(settings settings.Settings) (*gtk.Menu, error) {
	menu, err := gtk.MenuNew()
	if err != nil {
		log.Println(err)
	}

	app := item.App
	actions := app.GetActions()

	AddWindowsItemToMenu(menu, item.Windows, app)

	if item.Instances != 0 {
		separator, err := gtk.SeparatorMenuItemNew()
		if err == nil {
			menu.Append(separator)
		} else {
			log.Println(err)
		}
	}

	if actions != nil {
		for _, action := range actions {
			exec := func() {
				action.Run()
			}

			var actionMenuItem *gtk.MenuItem
			var err error

			if action.GetIcon() == "" {
				actionMenuItem, err = BuildContextItem(action.GetName(), exec)
			} else {
				actionMenuItem, err = BuildContextItem(action.GetName(), exec, action.GetIcon())
			}

			if err == nil {
				menu.Append(actionMenuItem)
			} else {
				log.Println(err)
			}
		}

		separator, err := gtk.SeparatorMenuItemNew()
		if err == nil {
			menu.Append(separator)
		} else {
			log.Println(err)
		}
	}

	launchMenuItem, err := BuildLaunchMenuItem(item)
	if err == nil {
		menu.Append(launchMenuItem)
	} else {
		log.Println(err)
	}

	pinMenuItem, err := BuildPinMenuItem(item, settings)
	if err == nil {
		menu.Append(pinMenuItem)
	} else {
		log.Println(err)
	}

	if item.Instances == 1 {
		closeMenuItem, err := BuildContextItem("Close", func() {
			ipc.Hyprctl("dispatch closewindow address:" + item.Windows[0]["Address"])
		}, "close-symbolic")
		if err == nil {
			menu.Append(closeMenuItem)
		} else {
			log.Println(err)
		}
	}

	menu.SetName("context-menu")
	menu.ShowAll()

	return menu, nil
}

func AddWindowsItemToMenu(menu *gtk.Menu, windows []map[string]string, app *desktop.App) {
	for _, window := range windows {
		menuItem, err := BuildContextItem(window["Title"], func() {
			go ipc.Hyprctl("dispatch focuswindow address:" + window["Address"])
		}, app.GetIcon())

		if err != nil {
			log.Println(err)
			continue
		}

		menu.Append(menuItem)
	}
}

func BuildLaunchMenuItem(item *Item) (*gtk.MenuItem, error) {
	app := item.App

	if item.Instances != 0 && app.GetSingleWindow() {
		return nil, errors.New("")
	}

	labelText := app.GetName()
	if item.Instances != 0 {
		labelText = "New Window - " + labelText
	}

	launchMenuItem, err := BuildContextItem(labelText, func() {
		app.Run()
	}, app.GetIcon())

	if err != nil {
		return nil, err
	}

	return launchMenuItem, nil
}

func BuildPinMenuItem(item *Item, settings settings.Settings) (*gtk.MenuItem, error) {
	labelText := "Pin"
	if item.IsPinned() {
		labelText = "Unpin"
	}

	menuItem, err := BuildContextItem(labelText, func() {
		item.TogglePin(settings)
	})

	if err != nil {
		return nil, err
	}

	return menuItem, nil
}

func BuildContextItem(labelText string, connectFunc func(), iconName ...string) (*gtk.MenuItem, error) {
	size := 16
	spacing := 6

	menuItem, err := gtk.MenuItemNew()
	if err != nil {
		return nil, err
	}

	menuItem.SetName("menu-item")

	hbox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, spacing)
	if err != nil {
		return nil, err
	}

	hbox.SetName("hbox")
	/* Hack (HELP ME)*/
	/* stackoverflow.com/questions/48452717/how-to-replace-the-deprecated-gtk3-gtkimagemenuitem */
	utils.AddStyle(hbox, fmt.Sprintf("#hbox {margin-left: %dpx;}", 0-(size+spacing)))

	label, err := gtk.LabelNew(labelText)
	if err != nil {
		return nil, err
	}

	label.SetEllipsize(pango.ELLIPSIZE_END)
	label.SetMaxWidthChars(30)

	if len(iconName) > 0 {
		icon, err := utils.CreateImage(iconName[0], size)
		if err == nil {
			hbox.Add(icon)
		}
	} else {
		label.SetMarginStart(size + spacing)
	}

	if connectFunc != nil {
		menuItem.Connect("activate", func() {
			connectFunc()
		})
	}

	hbox.Add(label)
	menuItem.SetReserveIndicator(false)
	menuItem.Add(hbox)

	return menuItem, nil
}
