package switcher

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/hysc"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/pkg/ipc"
)

// findIconPath searches for an icon file in common directories
func (s *Switcher) findIconPath(name string) string {
	if path, ok := s.iconPathCache[name]; ok {
		return path
	}

	if strings.Contains(name, "/") {
		if _, err := os.Stat(name); err == nil {
			s.iconPathCache[name] = name
			return name
		}
		return ""
	}

	// Common search paths
	searchPaths := []string{
		"/usr/share/icons/hicolor/48x48/apps",
		"/usr/share/icons/hicolor/scalable/apps",
		"/usr/share/icons/Adwaita/48x48/apps",
		"/usr/share/icons/Adwaita/scalable/apps",
		"/usr/share/pixmaps",
	}

	// Extensions
	exts := []string{".png", ".svg", ".xpm"}

	for _, dir := range searchPaths {
		for _, ext := range exts {
			path := filepath.Join(dir, name+ext)
			if _, err := os.Stat(path); err == nil {
				s.iconPathCache[name] = path
				return path // Found!
			}
			// Try lowercase
			pathLower := filepath.Join(dir, strings.ToLower(name)+ext)
			if _, err := os.Stat(pathLower); err == nil {
				s.iconPathCache[name] = pathLower
				return pathLower
			}
		}
	}
	s.iconPathCache[name] = "" // Mark as not found to avoid re-search
	return ""
}

// loadIconAsync loads an icon asynchronously and updates the image widget
func (s *Switcher) loadIconAsync(targetIcon *gtk.Image, name string, size int, currentGen int) {
	logTiming("[ICON] Starting async load for: %s", name)
	go func() {
		// 1. Try to find file manually (Thread Safe IO)
		path := s.findIconPath(name)

		if path != "" {
			// Load from file (Safe in BG)
			pixbuf, err := gdk.PixbufNewFromFileAtSize(path, size, size)
			if err == nil {
				glib.IdleAdd(func() {
					if s.renderGen != currentGen {
						return
					}
					targetIcon.SetFromPixbuf(pixbuf)
					logTiming("[ICON] Loaded from file: %s", name)
				})
				return
			}
		}

		// 2. Fallback: Load using Theme on Main Thread (Slow but Safe)
		glib.IdleAdd(func() {
			if s.renderGen != currentGen {
				return
			}
			theme, _ := gtk.IconThemeGetDefault()
			pixbuf, err := theme.LoadIcon(name, size, gtk.ICON_LOOKUP_FORCE_SIZE)
			if err == nil {
				targetIcon.SetFromPixbuf(pixbuf)
				logTiming("[ICON] Loaded from theme: %s", name)
			} else {
				// Final fallback
				targetIcon.SetFromIconName("application-default-icon", gtk.ICON_SIZE_DIALOG)
				targetIcon.SetPixelSize(size)
				logTiming("[ICON] Using fallback icon for: %s", name)
			}
		})
	}()
}

// capturePreviewAsync captures a window screenshot asynchronously
func (s *Switcher) capturePreviewAsync(
	client ipc.Client,
	scaledW, scaledH int,
	centerBox *gtk.Box,
	overlay *gtk.Overlay,
	initialIcon *gtk.Widget,
	iconName string,
	currentGen int,
	sem chan struct{},
	fingerprint string, // Window fingerprint for caching
) {
	stream, err := hysc.StreamNew(client.Address)
	if err != nil {
		return
	}

	// Configure stream BEFORE capture so scaling/masks are applied
	stream.SetFixedSize(scaledW, scaledH)
	stream.SetBorderRadius(4)
	stream.OnReady(func(sz *hysc.Size) {
		// Cache the screenshot pixbuf with timestamp for future use
		if pixbuf := stream.GetPixbuf(); pixbuf != nil {
			s.screenshotCache[fingerprint] = &CachedScreenshot{
				Pixbuf:    pixbuf,
				Timestamp: time.Now(),
			}
			logTiming("[SCREENSHOT] Cached screenshot for: %s", client.Address)
		}
	})

	logTiming("[SCREENSHOT] Starting async capture for: %s", client.Address)
	go func() {
		// Limit concurrency
		sem <- struct{}{}
		defer func() { <-sem }()

		// Capture using shared app
		var err error
		// Force fresh connection for each capture to resolve issue with multiple windows
		// if s.app != nil {
		// 	err = stream.CaptureFrameWithApp(s.app)
		// } else {
		err = stream.CaptureFrame()
		// }

		if err == nil {
			// Success! Swap on Main Thread
			glib.IdleAdd(func() {
				if s.renderGen != currentGen {
					return
				}

				centerBox.Remove(initialIcon)
				centerBox.Add(stream)

				// Add App Icon Badge
				badgeSize := 32
				if s.config.IconSize > 0 {
					badgeSize = s.config.IconSize
				} else {
					// Auto Mode
					badgeSize = scaledW / 8
					if badgeSize < 32 {
						badgeSize = 32
					}
					if badgeSize > 96 {
						badgeSize = 96
					}
				}
				badge, _ := utils.CreateImage(iconName, badgeSize)
				badge.SetHAlign(gtk.ALIGN_CENTER)
				badge.SetVAlign(gtk.ALIGN_START)
				badge.SetMarginTop(8)

				overlay.AddOverlay(badge)
				overlay.ShowAll()
				logTiming("[SCREENSHOT] Captured successfully for: %s", client.Address)
			})
		} else {
			// Log error but keep Icon
			debugLog("Preview failed for %s: %v", client.Address, err)
		}
	}()
}

// updatePreviews refreshes all window previews
func (s *Switcher) updatePreviews() {
	s.renderGen++ // Invalidate previous (though render skipped, but good practice)
	currentGen := s.renderGen

	// Concurrency limit
	sem := make(chan struct{}, 1)

	// Correct Approach: Traversal in Main Thread
	for i, w := range s.widgets {
		if i >= len(s.clients) {
			break
		}
		c := s.clients[i]

		// 1. Find CenterBox (The container for the image)
		centerBox := findCenterBox(w)
		if centerBox == nil {
			continue
		}

		// 2. Refresh Stream
		go func(client ipc.Client, targetBox *gtk.Box) {
			sem <- struct{}{}
			defer func() { <-sem }()

			// Check gen
			if s.renderGen != currentGen {
				return
			}

			stream, err := hysc.StreamNew(client.Address)
			if err != nil {
				return
			}

			stream.SetFixedSize(s.config.PreviewWidth, int(float64(s.config.PreviewWidth)*0.6)) // Approx
			stream.SetBorderRadius(4)

			// Capture
			if s.app != nil {
				err = stream.CaptureFrameWithApp(s.app)
			} else {
				err = stream.CaptureFrame()
			}

			if err == nil {
				glib.IdleAdd(func() {
					if s.renderGen != currentGen {
						return
					}
					// Swap
					// Remove all children of targetBox
					children := targetBox.GetChildren()
					children.Foreach(func(item interface{}) {
						targetBox.Remove(item.(gtk.IWidget))
					})
					targetBox.Add(stream)
					targetBox.ShowAll()
				})
			}
		}(c, centerBox)
	}
}

// findCenterBox finds the center box widget in the window widget hierarchy
func findCenterBox(root *gtk.Widget) *gtk.Box {
	// Root is winBox (Box)
	// Child 0 is Overlay
	// Overlay Child 0 is CenterBox

	// Cast to Container
	cWidget, _ := root.Cast()
	c, ok := cWidget.(*gtk.Container)
	if !ok {
		return nil
	}

	list := c.GetChildren()
	if list.Length() == 0 {
		return nil
	}

	// Overlay
	overlayWidget := list.Data().(*gtk.Widget)
	oWidget, _ := overlayWidget.Cast()
	overlay, ok := oWidget.(*gtk.Container)
	if !ok {
		return nil
	}

	// Overlay children
	olist := overlay.GetChildren()
	if olist.Length() == 0 {
		return nil
	}

	// CenterBox (First child)
	centerWidget := olist.Data().(*gtk.Widget)
	cw, _ := centerWidget.Cast()
	centerBox, ok := cw.(*gtk.Box)
	if !ok {
		return nil
	}

	return centerBox
}

// ToContainer safely casts a Widget to Container if possible
func ToContainer(w *gtk.Widget) *gtk.Container {
	// Attempt cast
	cWidget, _ := w.Cast()
	c, ok := cWidget.(*gtk.Container)
	if ok {
		return c
	}
	return nil
}
