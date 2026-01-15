package switcher

import (
	"fmt"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/internal/pkg/utils"
)

// initWindow creates and configures the main window
func (s *Switcher) initWindow() error {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return fmt.Errorf("unable to create window: %w", err)
	}
	s.window = win

	// Setup layer shell
	s.setupLayerShell()

	// Calculate and apply margins
	s.calculateAndApplyMargins()

	// Apply styles
	s.applyStyles()

	return nil
}

// setupLayerShell configures the layer shell for the window
func (s *Switcher) setupLayerShell() {
	layershell.InitForWindow(s.window)
	layershell.SetNamespace(s.window, "hypr-dock-switcher")
	layershell.SetLayer(s.window, layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetKeyboardMode(s.window, layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)
	layershell.SetExclusiveZone(s.window, -1)
	layershell.SetAnchor(s.window, layershell.LAYER_SHELL_EDGE_TOP, true)
	layershell.SetAnchor(s.window, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetAnchor(s.window, layershell.LAYER_SHELL_EDGE_LEFT, true)
	layershell.SetAnchor(s.window, layershell.LAYER_SHELL_EDGE_RIGHT, true)
}

// calculateAndApplyMargins calculates screen size and applies margins
func (s *Switcher) calculateAndApplyMargins() {
	// Custom Size via Margins
	scrW, scrH := 1920, 1080 // Safe default
	foundMonitor := false

	if display, err := gdk.DisplayGetDefault(); err == nil {
		monitor, err := display.GetPrimaryMonitor()
		if err != nil {
			monitor, err = display.GetMonitor(0)
		}

		if err == nil && monitor != nil {
			geo := monitor.GetGeometry()
			scrW = geo.GetWidth()
			scrH = geo.GetHeight()
			foundMonitor = true
		}
	}

	marginH := (scrW * (100 - s.config.WidthPercent)) / 200
	marginV := (scrH * (100 - s.config.HeightPercent)) / 200

	debugLog("Screen: %dx%d (Found: %v)", scrW, scrH, foundMonitor)
	debugLog("Margins: H=%d, V=%d", marginH, marginV)

	layershell.SetMargin(s.window, layershell.LAYER_SHELL_EDGE_LEFT, marginH)
	layershell.SetMargin(s.window, layershell.LAYER_SHELL_EDGE_RIGHT, marginH)
	layershell.SetMargin(s.window, layershell.LAYER_SHELL_EDGE_TOP, marginV)
	layershell.SetMargin(s.window, layershell.LAYER_SHELL_EDGE_BOTTOM, marginV)
}

// applyStyles applies CSS styles to the window and UI elements
func (s *Switcher) applyStyles() {
	// Make window transparent/dimmed
	utils.AddStyle(s.window, "window { background-color: rgba(20, 20, 20, 0.85); }")

	// Apply Font Size Globally via Screen Provider
	if cssProvider, err := gtk.CssProviderNew(); err == nil {
		css := fmt.Sprintf(`
        * { font-family: Sans; font-size: %dpx; }
        .switcher-item {
            background-color: rgba(40, 40, 40, 0.6);
            border: 4px solid #555;
            border-radius: 6px;
            transition: all 0.2s ease;
        }
        .switcher-item-selected {
            background-color: rgba(255, 255, 255, 0.2);
            border: 4px solid #ffffff;
            border-radius: 6px;
            box-shadow: 0 0 15px rgba(255, 255, 255, 0.4);
        }
        `, s.config.FontSize)
		cssProvider.LoadFromData(css)
		if screen, err := gdk.ScreenGetDefault(); err == nil {
			gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)
		}
	}
}

// createMainLayout creates the main UI layout with scrolling and debug box
func (s *Switcher) createMainLayout() {
	// Main Scrolled Window in case of many workspaces
	scroll, _ := gtk.ScrolledWindowNew(nil, nil)

	// Debug Indicator (Tiny box at top-left)
	s.debugBox, _ = gtk.EventBoxNew()
	s.debugBox.SetSizeRequest(20, 20)
	s.debugBox.SetHAlign(gtk.ALIGN_START)
	s.debugBox.SetVAlign(gtk.ALIGN_START)
	utils.AddStyle(s.debugBox, "eventbox { background-color: red; }")

	// Use Overlay to put Debug Box on top
	mainOverlay, _ := gtk.OverlayNew()
	mainOverlay.Add(scroll)
	mainOverlay.AddOverlay(s.debugBox)
	s.window.Add(mainOverlay)

	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_NEVER)

	// Main container (Vertical list of rows)
	s.box, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 30)
	s.box.SetHAlign(gtk.ALIGN_CENTER)
	s.box.SetVAlign(gtk.ALIGN_CENTER)
	s.box.SetMarginStart(50)
	s.box.SetMarginEnd(50)
	scroll.Add(s.box)
}
