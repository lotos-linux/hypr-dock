package switcher

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"

	"hypr-dock/internal/hysc"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/pkg/ipc"
)

const GridCols = 4

// render builds the UI for all workspaces and windows
func (s *Switcher) render() {
	// Increment generation to invalidate previous async tasks
	s.renderGen++
	currentGen := s.renderGen

	// Clear existing widgets
	children := s.box.GetChildren()
	children.Foreach(func(item interface{}) {
		s.box.Remove(item.(gtk.IWidget))
	})

	// Render a "Card" for each workspace
	// Serialize Wayland access (Panic Fix)
	sem := make(chan struct{}, 1)
	var currentRow *gtk.Box
	itemsInRow := 0

	for _, wsID := range s.workspaces {
		indices := s.workspaceMap[wsID]
		if len(indices) == 0 {
			continue
		}

		if itemsInRow%GridCols == 0 {
			currentRow, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 30)
			s.box.Add(currentRow)
		}
		itemsInRow++

		// Create workspace card
		s.createWorkspaceCard(wsID, indices, currentRow, currentGen, sem)
	}

	s.updateSelection()
	s.box.ShowAll()
}

// createWorkspaceCard creates a card for a single workspace
func (s *Switcher) createWorkspaceCard(wsID int, indices []int, currentRow *gtk.Box, currentGen int, sem chan struct{}) {
	// Find relevant monitor for this workspace
	firstClient := s.clients[indices[0]]
	mon, ok := s.monitorMap[firstClient.Monitor]
	if !ok && len(s.monitors) > 0 {
		mon = s.monitors[0]
	}

	// Calculate scale
	cardWidth := float64(s.config.PreviewWidth)
	if mon.Width <= 0 {
		mon.Width = 1920 // Fallback
	}
	scale := cardWidth / float64(mon.Width)
	cardHeight := float64(mon.Height) * scale

	// Workspace Container
	wsBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)

	// Header Box (Workspace info + Title)
	headerBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	headerBox.SetHAlign(gtk.ALIGN_FILL)

	// 1. Workspace Pill
	wsNumBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	utils.AddStyle(wsNumBox, "box { background-color: #333333; border-radius: 12px; padding: 2px 8px; }")

	wsNumLabel, _ := gtk.LabelNew(fmt.Sprintf("%d", wsID))
	utils.AddStyle(wsNumLabel, "label { color: #ffffff; font-weight: bold; font-size: 0.6em; }")
	wsNumBox.Add(wsNumLabel)
	headerBox.PackStart(wsNumBox, false, false, 0)

	// 2. Window Title
	titleText := firstClient.Title
	if titleText == "" {
		titleText = firstClient.Class
	}
	titleLabel, _ := gtk.LabelNew(titleText)
	titleLabel.SetHAlign(gtk.ALIGN_START)
	titleLabel.SetEllipsize(pango.ELLIPSIZE_END)
	titleLabel.SetMaxWidthChars(25)
	utils.AddStyle(titleLabel, "label { color: #aaaaaa; font-size: 0.7em; margin-left: 5px; }")
	headerBox.PackStart(titleLabel, true, true, 0)

	wsBox.Add(headerBox)

	// Fixed layout area representing the desktop
	fixed, _ := gtk.FixedNew()
	fixed.SetSizeRequest(int(cardWidth), int(cardHeight))
	utils.AddStyle(fixed, "fixed { background-color: rgba(60, 60, 60, 0.4); border-radius: 8px; }")
	wsBox.Add(fixed)
	currentRow.Add(wsBox)

	// Place Windows
	for _, idx := range indices {
		s.createWindowWidget(idx, mon, scale, fixed, currentGen, sem)
	}
}

// createWindowWidget creates a widget for a single window
func (s *Switcher) createWindowWidget(idx int, mon ipc.Monitor, scale float64, fixed *gtk.Fixed, currentGen int, sem chan struct{}) {
	c := s.clients[idx]

	// Relative coordinates
	relX := float64(c.At[0] - mon.X)
	relY := float64(c.At[1] - mon.Y)
	w := float64(c.Size[0])
	h := float64(c.Size[1])

	scaledX := int(relX * scale)
	scaledY := int(relY * scale)
	scaledW := int(w * scale)
	scaledH := int(h * scale)

	// Min size for visibility
	if scaledW < 24 {
		scaledW = 24
	}
	if scaledH < 24 {
		scaledH = 24
	}

	// Window Widget
	winBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	winBox.SetSizeRequest(scaledW, scaledH)

	// Apply base class
	ctx, _ := winBox.GetStyleContext()
	ctx.AddClass("switcher-item")

	// Container for content (Icon or Screenshot)
	overlay, _ := gtk.OverlayNew()
	winBox.PackStart(overlay, true, true, 0)

	// Background container (CenterBox)
	centerBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	centerBox.SetVAlign(gtk.ALIGN_CENTER)
	centerBox.SetHAlign(gtk.ALIGN_CENTER)
	overlay.Add(centerBox)

	// 1. Always Render Icon First
	iconName := c.Class
	iconSize := scaledW / 2
	if iconSize > 48 {
		iconSize = 48
	}
	if iconSize < 16 {
		iconSize = 16
	}

	// 2. Async Icon Load
	// Create Placeholder (Instant)
	icon, _ := gtk.ImageNewFromIconName("image-loading", gtk.ICON_SIZE_DIALOG)
	icon.SetPixelSize(iconSize)
	icon.SetHAlign(gtk.ALIGN_CENTER)
	centerBox.Add(icon)

	// Load icon asynchronously
	s.loadIconAsync(icon, iconName, iconSize, currentGen)

	// 3. Try to upgrade to Screenshot
	_, err := hysc.StreamNew(c.Address)
	if err == nil {
		// Check if we have a cached screenshot for this window
		fingerprint := getWindowFingerprint(c)
		if cachedPixbuf, exists := s.screenshotCache[fingerprint]; exists {
			// Reuse cached screenshot (instant!)
			logTiming("[SCREENSHOT] Using cached screenshot for: %s", c.Address)
			glib.IdleAdd(func() {
				// Check generation
				if s.renderGen != currentGen {
					return
				}

				// Replace icon with cached screenshot
				centerBox.Remove(icon)
				img, _ := gtk.ImageNewFromPixbuf(cachedPixbuf)
				img.SetHAlign(gtk.ALIGN_CENTER)
				centerBox.Add(img)
				overlay.ShowAll()
			})
		} else {
			// Capture new screenshot and cache it
			s.capturePreviewAsync(c, scaledW, scaledH, centerBox, overlay, icon.ToWidget(), iconName, currentGen, sem, fingerprint)
		}
	} else {
		log.Printf("Failed to create stream for %s: %v", c.Address, err)
	}

	// 4. Mouse Interaction (EventBox)
	eventBox, _ := gtk.EventBoxNew()
	eventBox.Add(winBox)

	// Click to confirm
	eventBox.Connect("button-press-event", s.createWindowClickHandler())

	// Hover to select
	eventBox.Connect("enter-notify-event", s.createWindowHoverHandler(idx))

	fixed.Put(eventBox, scaledX, scaledY)
	s.widgets[idx] = &winBox.Widget
}

// updateSelection updates the visual selection state of all widgets
func (s *Switcher) updateSelection() {
	for i, w := range s.widgets {
		if w == nil {
			continue
		}

		highlight := (i == s.selected)

		ctx, _ := w.GetStyleContext()
		if highlight {
			ctx.AddClass("switcher-item-selected")
		} else {
			ctx.RemoveClass("switcher-item-selected")
		}
	}

	// Ensure scrolling (basic approximation: scroll to selected)
	// Improving scroll logic is hard without coordinates, assume gtk handles it mostly or fixed size.
}
