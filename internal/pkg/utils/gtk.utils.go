package utils

import (
	"log"
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
	// Create image in file
	if strings.Contains(source, "/") {
		pixbuf, err := gdk.PixbufNewFromFileAtSize(source, size, size)
		if err != nil {
			log.Println(err)
			return CreateImage("image-missing", size)
		}

		return CreateImageFromPixbuf(pixbuf), nil
	}

	// Create image in icon name
	iconTheme, err := gtk.IconThemeGetDefault()
	if err != nil {
		log.Println("Unable to icon theme:", err)
		return CreateImage("image-missing", size)
	}

	pixbuf, err := iconTheme.LoadIcon(source, size, gtk.ICON_LOOKUP_FORCE_SIZE)
	if err != nil {
		log.Println(source, err)
		return CreateImage("image-missing", size)
	}

	return CreateImageFromPixbuf(pixbuf), nil
}

func CreateImageFromPixbuf(pixbuf *gdk.Pixbuf) *gtk.Image {
	image, err := gtk.ImageNewFromPixbuf(pixbuf)
	if err != nil {
		log.Println("Error creating image from pixbuf:", err)
		return nil
	}
	return image
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
		log.Printf("Failed to create CSS provider: %v", err)
		return errors.Wrap(err, "failed to create CSS provider")
	}

	if err := cssProvider.LoadFromPath(cssFile); err != nil {
		log.Printf("Failed to load CSS from %q: %v", cssFile, err)
		return errors.Wrapf(err, "failed to load CSS from %q", cssFile)
	}

	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		log.Printf("Failed to get default screen: %v", err)
		return errors.Wrap(err, "failed to get default screen")
	}

	gtk.AddProviderForScreen(
		screen, cssProvider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	return nil
}

func RemoveStyleProvider(widget *gtk.Box, provider *gtk.CssProvider) {
	if provider == nil {
		log.Println("provider is nil")
		return
	}

	styleContext, err := widget.GetStyleContext()
	if err != nil {
		log.Println(err)
		return
	}

	styleContext.RemoveProvider(provider)
}
