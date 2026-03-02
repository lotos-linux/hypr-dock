package item

import (
	"errors"
	"fmt"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/hashicorp/go-hclog"

	"hypr-dock/internal/desktop"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/pkg/ipc"
)

func (i *Item) WindowsMenu() (*gtk.Menu, error) {
	menu, err := gtk.MenuNew()
	if err != nil {
		return nil, err
	}

	AddWindowsItemToMenu(menu, i.Windows, i.App, i.log)

	menu.SetName("windows-menu")
	menu.ShowAll()

	return menu, nil
}

func (i *Item) ContextMenu() (*gtk.Menu, error) {
	menu, err := gtk.MenuNew()
	if err != nil {
		return nil, err
	}

	app := i.App
	actions := app.GetActions()

	AddWindowsItemToMenu(menu, i.Windows, app, i.log)

	if len(i.Windows) != 0 {
		separator, err := gtk.SeparatorMenuItemNew()
		if err == nil {
			menu.Append(separator)
		} else {
			i.log.Error("Unable to create gtk separator", "error", err)
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
				i.log.Error("Unable to create context item", "error", err)
			}
		}

		separator, err := gtk.SeparatorMenuItemNew()
		if err == nil {
			menu.Append(separator)
		} else {
			i.log.Error("Unable to create gtk separator", "error", err)
		}
	}

	launchMenuItem, err := BuildLaunchMenuItem(i)
	if err == nil {
		menu.Append(launchMenuItem)
	} else {
		i.log.Error("Unable to create launch menu item", "error", err)
	}

	pinMenuItem, err := BuildPinMenuItem(i)
	if err == nil {
		menu.Append(pinMenuItem)
	} else {
		i.log.Error("Unable to create pin menu item", "error", err)
	}

	if len(i.Windows) == 1 {
		client, ok := utils.GetSingleValue(i.Windows)
		if ok {
			closeMenuItem, err := BuildContextItem("Close", func() {
				ipc.Hyprctl("dispatch closewindow address:" + client.Address)
			}, "close-symbolic")
			if err == nil {
				menu.Append(closeMenuItem)
			} else {
				i.log.Error("Unable to create close menu item", "error", err)
			}
		}
	}

	menu.SetName("context-menu")
	menu.ShowAll()

	return menu, nil
}

func AddWindowsItemToMenu(menu *gtk.Menu, windows map[string]*ipc.Client, app *desktop.App, log hclog.Logger) {
	for _, window := range windows {
		menuItem, err := BuildContextItem(window.Title, func() {
			go ipc.Hyprctl("dispatch focuswindow address:" + window.Address)
		}, app.GetIcon())

		if err != nil {
			log.Error("Unable to create launch menu item", "error", err)
			continue
		}

		menu.Append(menuItem)
	}
}

func BuildLaunchMenuItem(item *Item) (*gtk.MenuItem, error) {
	app := item.App

	instances := len(item.Windows)

	if instances != 0 && app.GetSingleWindow() {
		return nil, errors.New("")
	}

	labelText := app.GetName()
	if instances != 0 {
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

func BuildPinMenuItem(item *Item) (*gtk.MenuItem, error) {
	labelText := "Pin"
	if item.IsPinned() {
		labelText = "Unpin"
	}

	menuItem, err := BuildContextItem(labelText, func() {
		item.TogglePin()
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
		icon, err := utils.CreateImage(iconName[0], size, hbox)
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
