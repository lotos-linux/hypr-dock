package utils

import (
	"math"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pkg/errors"
)

func CreateImageWidthTransform(source string, size int, parent gtk.IWidget, scaleFactor float64, rotate bool) (*gtk.Image, error) {
	scaleSize := int(math.Round(float64(size) * math.Max(scaleFactor, 0)))

	return CreateImage0(source, scaleSize, parent, rotate)
}

func CreateImage(source string, size int, parent gtk.IWidget) (*gtk.Image, error) {
	image, err := CreateImage0(source, size, parent, false)
	if err == nil {
		return image, nil
	}
	return CreateImage("image-missing", size, parent)
}

func CreateImage0(source string, size int, parent gtk.IWidget, rotate bool) (*gtk.Image, error) {
	w := parent.ToWidget()
	var err error
	var pixbuf *gdk.Pixbuf

	scaleFactor := w.GetScaleFactor()
	physicalSize := size * scaleFactor

	if strings.Contains(source, "/") {
		pixbuf, err = gdk.PixbufNewFromFileAtSize(source, physicalSize, physicalSize)
	} else {
		theme, _ := gtk.IconThemeGetDefault()
		pixbuf, err = theme.LoadIcon(source, physicalSize, gtk.ICON_LOOKUP_FORCE_SIZE)
	}
	if err != nil {
		return nil, err
	}

	if rotate {
		rotPixbuf, err := pixbuf.RotateSimple(gdk.PIXBUF_ROTATE_COUNTERCLOCKWISE)
		if err == nil {
			pixbuf = rotPixbuf
		}
	}

	surface, err := gdk.CairoSurfaceCreateFromPixbuf(pixbuf, scaleFactor, nil)
	if err != nil {
		return nil, err
	}

	image, err := gtk.ImageNew()
	if err != nil {
		return nil, err
	}

	image.SetFromSurface(surface)
	image.SetPixelSize(size)

	return image, nil
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
