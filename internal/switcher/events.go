package switcher

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// setupEventHandlers configures all keyboard and mouse event handlers
func (s *Switcher) setupEventHandlers() {
	// Keyboard events
	s.window.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		s.handleKeyPress(ev)
	})

	s.window.Connect("key-release-event", func(win *gtk.Window, ev *gdk.Event) {
		s.handleKeyRelease(ev)
	})

	// Window destroy
	s.window.Connect("destroy", func() {
		if s.app != nil {
			s.app.Close()
		}
		gtk.MainQuit()
	})
}

// handleKeyPress processes keyboard press events
func (s *Switcher) handleKeyPress(ev *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(ev)
	keyVal := keyEvent.KeyVal()

	switch keyVal {
	case gdk.KEY_Tab, gdk.KEY_ISO_Left_Tab:
		if gdk.ModifierType(keyEvent.State())&gdk.SHIFT_MASK != 0 {
			s.cycle(-1)
		} else {
			s.cycle(1)
		}
	case gdk.KEY_Left:
		s.cycle(-1)
	case gdk.KEY_Right:
		s.cycle(1)
	case gdk.KEY_Up:
		s.moveGrid(-1)
	case gdk.KEY_Down:
		s.moveGrid(1)
	case gdk.KEY_Return, gdk.KEY_KP_Enter:
		s.confirm()
	case gdk.KEY_Escape:
		s.visible = false
		s.window.Hide()
		// gtk.MainQuit() // Daemon mode: Hide instead
	}
}

// handleKeyRelease processes keyboard release events
func (s *Switcher) handleKeyRelease(ev *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(ev)
	keyVal := keyEvent.KeyVal()
	debugLog("Key Release: %d", keyVal)

	// Check for Alt (L/R) or Super (L/R)
	if keyVal == gdk.KEY_Alt_L || keyVal == gdk.KEY_Alt_R ||
		keyVal == gdk.KEY_Meta_L || keyVal == gdk.KEY_Meta_R ||
		keyVal == gdk.KEY_Super_L || keyVal == gdk.KEY_Super_R {

		// utils.AddStyle(s.debugBox, "eventbox { background-color: red; }")
		s.confirm()
	}
}

// createWindowHoverHandler creates a hover event handler for window selection
func (s *Switcher) createWindowHoverHandler(idx int) func() {
	return func() {
		s.selected = idx
		s.updateSelection()
	}
}

// createWindowClickHandler creates a click event handler for window confirmation
func (s *Switcher) createWindowClickHandler() func() {
	return func() {
		s.confirm()
	}
}
