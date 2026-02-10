package utils

import (
	"fmt"
	"math"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"
)

func CreateImageWidthScale(source string, size int, scaleFactor float64) (*gtk.Image, error) {
	scaleSize := int(math.Round(float64(size) * math.Max(scaleFactor, 0)))

	return CreateImage(source, scaleSize)
}

func CreateImage(source string, size int) (*gtk.Image, error) {
	pixbuf, err := CreatePixbuf(source, size)
	if err != nil {
		return nil, err
	}
	return gtk.ImageNewFromPixbuf(pixbuf)
}

func CreatePixbuf(source string, size int) (*gdk.Pixbuf, error) {
	// Create image in file
	if strings.Contains(source, "/") {
		pixbuf, err := gdk.PixbufNewFromFileAtSize(source, size, size)
		if err != nil {
			return CreatePixbuf("image-missing", size)
		}
		return pixbuf, nil
	}

	// Create image in icon name
	iconTheme, err := gtk.IconThemeGetDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to get default gtk icon theme: %v", err)
	}

	pixbuf, err := iconTheme.LoadIcon(source, size, gtk.ICON_LOOKUP_FORCE_SIZE)
	if err != nil {
		return CreatePixbuf("image-missing", size)
	}

	return pixbuf, nil
}

func AddStyle(widget gtk.IWidget, style string) (*gtk.CssProvider, error) {
	provider, err := gtk.CssProviderNew()
	if err != nil {
		return nil, err
	}

	err = provider.LoadFromData(style)
	if err != nil {
		return nil, err
	}

	context, err := widget.ToWidget().GetStyleContext()
	if err != nil {
		return nil, err
	}

	context.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	return provider, nil
}

func AddCssProvider(cssFile string) error {
	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		return errors.Wrap(err, "failed to create CSS provider")
	}

	if err := cssProvider.LoadFromPath(cssFile); err != nil {
		return errors.Wrapf(err, "failed to load CSS from %q", cssFile)
	}

	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		return errors.Wrap(err, "failed to get default screen")
	}

	gtk.AddProviderForScreen(
		screen, cssProvider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	return nil
}

func RemoveStyleProvider(widget *gtk.Box, provider *gtk.CssProvider) error {
	if provider == nil {
		return errors.New("provider is nil")
	}

	styleContext, err := widget.GetStyleContext()
	if err != nil {
		return err
	}

	styleContext.RemoveProvider(provider)
	return nil
}

func GetFirstAvailableImage(sources []string, fallback ...string) string {
	fallbackImg := "image-missing"
	if len(fallback) > 0 {
		fallbackImg = fallback[0]
	}

	theme, err := gtk.IconThemeGetDefault()
	if err != nil {
		return fallbackImg
	}

	for _, source := range sources {
		if strings.Contains(source, "/") && FileExists(source) {
			return source
		}

		if theme.HasIcon(source) {
			return source
		}
	}

	return fallbackImg
}

func SetCursorPointer(v *gtk.Widget) error {
	display, err := gdk.DisplayGetDefault()
	if err != nil {
		return err
	}

	pointer, _ := gdk.CursorNewFromName(display, "pointer")
	arrow, _ := gdk.CursorNewFromName(display, "default")

	v.Connect("enter-notify-event", func() {
		win, _ := v.GetWindow()
		if win != nil {
			win.SetCursor(pointer)
		}
	})

	v.Connect("leave-notify-event", func(_ interface{}, e *gdk.Event) {
		event := gdk.EventCrossingNewFromEvent(e)
		win, _ := v.GetWindow()

		if win != nil && event.Detail() != 2 {
			win.SetCursor(arrow)
		}
	})

	return nil
}

func SetAutoHover(v *gtk.Widget, context *gtk.StyleContext) {
	v.Connect("enter-notify-event", func() {
		context.AddClass("hover")
	})
	v.Connect("leave-notify-event", func(_ interface{}, e *gdk.Event) {
		event := gdk.EventCrossingNewFromEvent(e)
		isInWindow := event.Detail() == 3 || event.Detail() == 0

		if isInWindow {
			context.RemoveClass("hover")
		}
	})
}
