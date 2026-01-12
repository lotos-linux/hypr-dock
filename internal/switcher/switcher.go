package switcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"

	"hypr-dock/internal/hysc"
	"hypr-dock/internal/pkg/utils"
	"hypr-dock/pkg/ipc"
	"hypr-dock/pkg/wl"

	"github.com/hashicorp/go-hclog"
)

const GridCols = 4

type Switcher struct {
	window *gtk.Window
	box    *gtk.Box

	// Data
	clients      []ipc.Client
	workspaceMap map[int][]int // WorkspaceID -> Indices in s.clients
	workspaces   []int         // Sorted Workspace IDs

	// Layout
	monitors   []ipc.Monitor
	monitorMap map[int]ipc.Monitor

	// State
	selected   int           // Index in s.clients
	widgets    []*gtk.Widget // Map client index to its widget
	app        *wl.App       // Single wayland app connection
	debugBox   *gtk.EventBox // For visual debugging of Alt key
	altWasHeld bool          // State for polling
	startTime  time.Time     // Grace period
	config     Config        // Loaded config

	// Daemon State
	visible       bool
	iconPathCache map[string]string
	renderGen     int
}

func Run() {
	gtk.Init(nil)

	// Singleton Check
	s := &Switcher{
		selected:      0,
		workspaceMap:  make(map[int][]int),
		monitorMap:    make(map[int]ipc.Monitor),
		altWasHeld:    true, // ASSUME Held on start
		startTime:     time.Now(),
		config:        LoadConfig(),
		visible:       true,
		iconPathCache: make(map[string]string),
	}

	log.Printf("------- SWITCHER CONFIG -------")
	log.Printf("Font Size: %d", s.config.FontSize)
	log.Printf("Size: %d%% x %d%%", s.config.WidthPercent, s.config.HeightPercent)
	log.Printf("Preview Width: %dpx", s.config.PreviewWidth)
	log.Printf("Show All Monitors: %v", s.config.ShowAllMonitors)
	log.Printf("-------------------------------")

	if !setupSingleton(s) {
		return
	}
	defer os.Remove("/tmp/hypr-dock-switcher.lock")

	// Signal handling for Cycling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)
	go func() {
		for range sigChan {
			glib.IdleAdd(func() {
				s.handleSignal()
			})
		}
	}()

	app, err := wl.NewApp(hclog.Default())
	if err != nil {
		log.Printf("Failed to connect to Wayland: %v", err)
	} else {
		s.app = app
		defer app.Close()
	}

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	s.window = win

	layershell.InitForWindow(win)
	layershell.SetNamespace(win, "hypr-dock-switcher")
	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_OVERLAY)
	layershell.SetKeyboardMode(win, layershell.LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)
	layershell.SetExclusiveZone(win, -1)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_TOP, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, true)

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

	log.Printf("Screen: %dx%d (Found: %v)", scrW, scrH, foundMonitor)
	log.Printf("Margins: H=%d, V=%d", marginH, marginV)

	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, marginH)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, marginH)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, marginV)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, marginV)

	// Make window transparent/dimmed
	utils.AddStyle(win, "window { background-color: rgba(20, 20, 20, 0.85); }")

	// Apply Font Size Globally via Screen Provider
	if cssProvider, err := gtk.CssProviderNew(); err == nil {
		css := fmt.Sprintf(`
        * { font-family: Sans; font-size: %dpx; }
        .switcher-item {
            background-color: rgba(40, 40, 40, 0.6);
            border: 2px solid #555;
            border-radius: 6px;
            transition: all 0.2s ease;
        }
        .switcher-item-selected {
            background-color: rgba(50, 150, 255, 0.4);
            border: 2px solid #3296ff;
            border-radius: 6px;
            box-shadow: 0 0 10px rgba(50, 150, 255, 0.3);
        }
        `, s.config.FontSize)
		cssProvider.LoadFromData(css)
		if screen, err := gdk.ScreenGetDefault(); err == nil {
			gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)
		}
	}

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
	win.Add(mainOverlay)

	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_NEVER)
	// win.Add(scroll) // Removed because added to overlay

	// Main container (Vertical list of rows)
	s.box, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 30)
	s.box.SetHAlign(gtk.ALIGN_CENTER)
	s.box.SetVAlign(gtk.ALIGN_CENTER)
	s.box.SetMarginStart(50)
	s.box.SetMarginEnd(50)
	scroll.Add(s.box)

	s.loadData()
	s.render()

	win.ShowAll()
	win.Connect("destroy", func() {
		if s.app != nil {
			s.app.Close()
		}
		gtk.MainQuit()
	})

	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
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
	})

	// Detect Alt Release to confirm
	win.Connect("key-release-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(ev)
		keyVal := keyEvent.KeyVal()
		log.Printf("Key Release: %d", keyVal)

		// Check for Alt (L/R) or Super (L/R)
		if keyVal == gdk.KEY_Alt_L || keyVal == gdk.KEY_Alt_R ||
			keyVal == gdk.KEY_Meta_L || keyVal == gdk.KEY_Meta_R ||
			keyVal == gdk.KEY_Super_L || keyVal == gdk.KEY_Super_R {

			// utils.AddStyle(s.debugBox, "eventbox { background-color: red; }")
			s.confirm()
		}
	})

	win.SetKeepAbove(true)
	win.Present()
	win.ShowAll()

	// Initial selection logic moved to loadData (Smart History)

	// Polling for Alt State (every 50ms)
	glib.TimeoutAdd(50, s.checkModifiers)

	gtk.Main()
}

func (s *Switcher) loadData() {
	// 1. Get Monitors to know limits/offsets
	mons, err := ipc.GetMonitors()
	if err == nil {
		s.monitors = mons
		for _, m := range mons {
			s.monitorMap[m.Id] = m
		}
	}

	// 2. Get Clients
	all, err := ipc.GetClients()
	if err != nil {
		log.Println(err)
		return
	}

	// Filter & Group
	var filtered []ipc.Client
	s.workspaceMap = make(map[int][]int)

	// Determine Active Monitor if filtering
	activeMonitorID := -1
	if !s.config.ShowAllMonitors {
		for _, m := range s.monitors {
			if m.Focused {
				activeMonitorID = m.Id
				break
			}
		}
	}

	// Sort by Workspace ID then maybe processing order
	// We want a stable order for cycling.
	sort.Slice(all, func(i, j int) bool {
		if all[i].Workspace.Id != all[j].Workspace.Id {
			return all[i].Workspace.Id < all[j].Workspace.Id
		}
		return all[i].At[0] < all[j].At[0] // Left-to-right within workspace
	})

	for _, c := range all {
		if c.Mapped && c.Workspace.Id > 0 {
			// Check Monitor Filter
			if !s.config.ShowAllMonitors && activeMonitorID != -1 {
				if c.Monitor != activeMonitorID {
					continue
				}
			}

			// Check if monitor exists for this workspace (approximation)
			// Hyprland client has Monitor ID
			filtered = append(filtered, c)
			idx := len(filtered) - 1
			s.workspaceMap[c.Workspace.Id] = append(s.workspaceMap[c.Workspace.Id], idx)
		}
	}
	s.clients = filtered
	s.widgets = make([]*gtk.Widget, len(s.clients))

	// Get list of workspaces
	var ws []int
	for k := range s.workspaceMap {
		ws = append(ws, k)
	}
	sort.Ints(ws)
	s.workspaces = ws

	// Smart Selection: Find Previous Window (FocusHistoryID == 1)
	targetIdx := 0
	if len(s.clients) > 1 {
		// Default fallback if we can't find history
		targetIdx = 1
	}

	// Scan for FocusHistoryID == 1
	for i, c := range s.clients {
		if c.FocusHistoryID == 1 {
			targetIdx = i
			break
		}
	}

	// Safety: If for some reason we picked the current window (0) but others exist, force move.
	if targetIdx == 0 && len(s.clients) > 1 {
		targetIdx = 1
	}

	s.selected = targetIdx
	debugLog("DATA LOADED: %d Clients, %d Workspaces. Selected Index: %d. CycleWorkspaces=%v", len(s.clients), len(s.workspaces), s.selected, s.config.CycleWorkspaces)
}

func (s *Switcher) render() {
	// Increment generation to invalidate previous async tasks
	s.renderGen++
	currentGen := s.renderGen

	// Clear
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

		// Find relevant monitor for this workspace
		// We look at the first client's monitor ID
		firstClient := s.clients[indices[0]]
		mon, ok := s.monitorMap[firstClient.Monitor]
		if !ok && len(s.monitors) > 0 {
			mon = s.monitors[0]
		}

		// Calculate scale
		// Target width for workspace card: Configurable
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

			// Safe Async Load: Manual Search -> BG Thread; Fallback -> Main Thread
			go func(targetIcon *gtk.Image, name string, size int) {
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
					} else {
						// Final fallback
						targetIcon.SetFromIconName("application-default-icon", gtk.ICON_SIZE_DIALOG)
						targetIcon.SetPixelSize(size)
					}
				})
			}(icon, iconName, iconSize)

			// 2. Try to upgrade to Screenshot asynchronously
			stream, err := hysc.StreamNew(c.Address)
			if err == nil {
				// Configure stream BEFORE capture so scaling/masks are applied
				stream.SetFixedSize(scaledW, scaledH)
				stream.SetBorderRadius(4)
				stream.OnReady(func(sz *hysc.Size) {
					// Silent handler to avoid warnings
				})

				go func(stream *hysc.Stream, targetOffset *gtk.Box, targetOverlay *gtk.Overlay, initialIcon *gtk.Widget) {
					// Limit concurrency
					sem <- struct{}{}
					defer func() { <-sem }()

					// Capture using shared app
					var err error
					if s.app != nil {
						err = stream.CaptureFrameWithApp(s.app)
					} else {
						err = stream.CaptureFrame() // Fallback (shouldn't happen if connected)
					}

					if err == nil {
						// Success! Swap on Main Thread
						glib.IdleAdd(func() {
							if s.renderGen != currentGen {
								return
							}

							targetOffset.Remove(initialIcon)
							targetOffset.Add(stream)

							// Add App Icon Badge
							badgeSize := 32
							if scaledW < 100 {
								badgeSize = 16
							}
							badge, _ := utils.CreateImage(iconName, badgeSize)
							badge.SetHAlign(gtk.ALIGN_CENTER)
							badge.SetVAlign(gtk.ALIGN_START)
							badge.SetMarginTop(8)

							targetOverlay.AddOverlay(badge)
							targetOverlay.ShowAll()
						})
					} else {
						// Log error but keep Icon
						log.Printf("Preview failed for %s: %v", c.Address, err)
					}
				}(stream, centerBox, overlay, icon.ToWidget())
			}

			// 3. Mouse Interaction (EventBox)
			eventBox, _ := gtk.EventBoxNew()
			eventBox.Add(winBox)

			// Click to confirm
			eventBox.Connect("button-press-event", func() {
				s.confirm()
			})

			// Hover to select
			// Capture idx for closure
			currentIdx := idx
			eventBox.Connect("enter-notify-event", func() {
				s.selected = currentIdx
				s.updateSelection()
			})

			fixed.Put(eventBox, scaledX, scaledY)
			s.widgets[idx] = &winBox.Widget
		}
	}

	s.updateSelection()
	s.box.ShowAll()
}

func (s *Switcher) updateSelection() {
	// var selectedWorkspaceID int
	// if len(s.clients) > 0 {
	// 	selectedWorkspaceID = s.clients[s.selected].Workspace.Id
	// }

	for i, w := range s.widgets {
		if w == nil {
			continue
		}

		highlight := (i == s.selected)

		// If Cycle Workspaces, highlight if in same workspace
		// BUT only if we have multiple workspaces, otherwise it looks static (everything highlighted)
		// FIXED: User finds this confusing ("static"). Let's disable "Workspace Highlighting" visually
		// and always highlight just the selected window. The cycling logic still jumps workspaces,
		// but the visual feedback is precise.
		/*
			if s.config.CycleWorkspaces && len(s.workspaces) > 1 {
				if s.clients[i].Workspace.Id == selectedWorkspaceID {
					highlight = true
				}
			}
		*/

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

func (s *Switcher) cycle(direction int) {
	if len(s.clients) == 0 {
		return
	}

	if s.config.CycleWorkspaces {
		// Jump to next workspace
		currentID := s.clients[s.selected].Workspace.Id
		nextIdx := s.selected
		foundNewWorkspace := false

		// Limit loop to length to avoid infinite
		for i := 0; i < len(s.clients); i++ {
			nextIdx += direction

			// Wrap
			if nextIdx >= len(s.clients) {
				nextIdx = 0
			}
			if nextIdx < 0 {
				nextIdx = len(s.clients) - 1
			}

			if s.clients[nextIdx].Workspace.Id != currentID {
				// Found new workspace
				s.selected = nextIdx
				foundNewWorkspace = true
				break
			}
		}

		// Fallback: If we couldn't find a different workspace (e.g. only 1 workspace open),
		// treat it as normal window cycling so the user isn't stuck.
		if !foundNewWorkspace {
			debugLog("CYCLE: No new workspace found. Fallback to window cycling.")
			s.selected += direction
			if s.selected >= len(s.clients) {
				s.selected = 0
			}
			if s.selected < 0 {
				s.selected = len(s.clients) - 1
			}
		} else {
			debugLog("CYCLE: Switched to workspaceID %d (Client Idx %d)", s.clients[s.selected].Workspace.Id, s.selected)
		}
	} else {
		// Standard Window Cycling
		s.selected += direction
		if s.selected >= len(s.clients) {
			s.selected = 0
		}
		if s.selected < 0 {
			s.selected = len(s.clients) - 1
		}
		debugLog("CYCLE: Window Index %d", s.selected)
	}

	s.updateSelection()
}

func (s *Switcher) confirm() {
	if len(s.clients) > 0 && s.selected < len(s.clients) {
		selectedClient := s.clients[s.selected]
		addr := selectedClient.Address

		fmt.Printf("Switching to %s (Workspace: %d)\n", selectedClient.Title, selectedClient.Workspace.Id)
		debugLog("CONFIRM: Switching to %s", selectedClient.Address)

		// Hide Immediately
		s.visible = false
		s.window.Hide()

		// Async Call
		go func() {
			ipc.Hyprctl(fmt.Sprintf("dispatch focuswindow address:%s", addr))
		}()
	} else {
		s.visible = false
		s.window.Hide()
	}
}

func (s *Switcher) handleSignal() {
	debugLog("SIGNAL: Received SIGUSR1. Visible=%v", s.visible)
	if !s.visible {
		// Wake up!
		s.visible = true
		s.startTime = time.Now()
		s.altWasHeld = true

		// Refresh Data
		// oldClients := s.clients
		s.loadData()

		// Check if we can reuse the existing UI (Instant Show)
		// FIXED: Users report stale screenshots. Always re-render to ensure fresh previews.
		// if clientsEqual(oldClients, s.clients) {
		// 	debugLog("CACHE: Windows unchanged, skipping render.")
		// 	s.updateSelection()
		// 	// s.updatePreviews()
		// } else {
		// debugLog("CACHE: Windows changed (Old: %d, New: %d), re-rendering.", len(oldClients), len(s.clients))
		s.render()
		// }

		s.window.SetKeepAbove(true)
		s.window.Present()
		s.window.Activate()
		s.window.ShowAll()

		// Restart polling
		glib.TimeoutAdd(50, s.checkModifiers)
	} else {
		// Already visible, treat as "Cycle Next" (User hit Alt+Tab again while we were open)
		s.cycle(1)
	}
}

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
func (s *Switcher) moveGrid(rowDelta int) {
	if len(s.clients) == 0 {
		return
	}

	// 1. Identify current workspace
	currentClient := s.clients[s.selected]
	currentWsID := currentClient.Workspace.Id

	// 2. Find index in s.workspaces
	currentWsIdx := -1
	for i, wsID := range s.workspaces {
		if wsID == currentWsID {
			currentWsIdx = i
			break
		}
	}

	if currentWsIdx == -1 {
		return
	}

	// 3. Calculate target workspace index
	targetWsIdx := currentWsIdx + (rowDelta * GridCols)

	// Clamp/Wrap
	if targetWsIdx < 0 {
		targetWsIdx = len(s.workspaces) + targetWsIdx
		if targetWsIdx < 0 {
			targetWsIdx = 0
		}
	} else if targetWsIdx >= len(s.workspaces) {
		targetWsIdx = targetWsIdx % len(s.workspaces)
	}

	// 4. Select first client of target workspace
	targetWsID := s.workspaces[targetWsIdx]
	indices := s.workspaceMap[targetWsID]
	if len(indices) > 0 {
		s.selected = indices[0]
		s.updateSelection()
	}
}

func (s *Switcher) checkModifiers() bool {
	if !s.visible {
		return false // Stop polling if hidden
	}
	if s.window == nil {
		return true
	}
	gdkWin, _ := s.window.GetWindow()
	if gdkWin == nil {
		return true
	}

	display, _ := s.window.GetDisplay()
	seat, _ := display.GetDefaultSeat()
	pointer, _ := seat.GetPointer()
	_, _, _, state := gdkWin.GetDevicePosition(pointer)

	// Check for Alt (Mod1) or Super (Mod4)
	isAltDown := (state&gdk.MOD1_MASK != 0) || (state&gdk.SUPER_MASK != 0) || (state&gdk.META_MASK != 0)

	// DEBUG: Explicitly log state every ~250ms to avoid spam but show status
	// Or just log on change?
	// For "detailed", let's log every check that has modifiers.
	if isAltDown || s.altWasHeld {
		// debugLog("POLL: Mods=%v AltDown=%v Held=%v TimeSinceStart=%v", state, isAltDown, s.altWasHeld, time.Since(s.startTime))
	}

	// Grace Period: 150ms (Balanced for reliability vs speed)
	if time.Since(s.startTime) < 150*time.Millisecond {
		// Just update visual if we DO see it
		if isAltDown {
			s.altWasHeld = true
			if s.debugBox != nil {
				// Avoid AddStyle in loop (memory leak)
				// status update only if changed? ignoring for now to fix freeze.
			}
			debugLog("POLL [Grace]: saw alt down. Held=true")
			return true
		}
		debugLog("POLL [Grace]: alt NOT down. Ignoring release (Grace period).")
		// If NOT held, and we previously saw it held, it's a fast tap!
		// Don't return true, fall through to release logic below.
	}

	if isAltDown {
		if !s.altWasHeld {
			debugLog("POLL: Alt pressed (Transition to Held)")
		}
		s.altWasHeld = true
		// utils.AddStyle(s.debugBox, "eventbox { background-color: #00ff00; }") // Green
	} else {
		// If we thought it was held, and now it's NOT held, then confirm
		if s.altWasHeld {
			debugLog("POLL: Alt Released! Confirming...")
			// utils.AddStyle(s.debugBox, "eventbox { background-color: red; }") // Red
			s.confirm()
			return false // Stop polling
		}
	}

	return true
}

func setupSingleton(s *Switcher) bool {
	lockFile := "/tmp/hypr-dock-switcher.lock"

	// Check if lock file exists
	if data, err := ioutil.ReadFile(lockFile); err == nil {
		pid, err := strconv.Atoi(string(data))
		if err == nil {
			// Check if process is still alive
			process, err := os.FindProcess(pid)
			if err == nil {
				// Send SIGUSR1 to cycle
				if err := process.Signal(syscall.SIGUSR1); err == nil {
					// We are the client CLI, so we exit
					// But we can't easily use debugLog unless we define it... (we did globally)
					// Wait, this runs in the NEW process. It has access to debugLog too.
					debugLog("CLIENT: Sending SIGUSR1 to existing PID %d", pid)
					return false // Signal sent, exit this instance
				}
			}
		}
	}

	// Write current PID
	pid := os.Getpid()
	_ = ioutil.WriteFile(lockFile, []byte(strconv.Itoa(pid)), 0644)
	debugLog("STARTUP: Daemon Started. PID=%d", pid)
	return true
}

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
		// Widget -> Box
		// Access internal pointer logic is hard in gotk3 without casting.
		// Helper:
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

			// Calc size (Approx or stored? We can use fixed size or query?)
			// We configured stream in render. We should ideally reuse size.
			// Let's use 300x200 default or try to match config?
			// Render used scaledW, scaledH. We lost that context.
			// BUT, stream.SetFixedSize is optional? No, crucial for scaling.
			// Re-calculating scale is annoying.
			// Maybe just capture full and let GTK scale? (SetFixedSize does capture scaling).
			// Let's use PREVIEW_WIDTH/2 approx?

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

// Helper to find the CenterBox (Image Container)
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

// Type casting helper needed?
// gotk3: root.Cast() returns *Object. We need type assertion on the correct wrapper.
// Usually: obj := root.ToContainer() if compatible?
// Let's use a simpler "ToContainer" map if defined? No.

// Better helper using GLib type system checking
// Helper to safely cast a Widget to Container if possible
// Since gotk3 doesn't have a direct "IsContainer" check on Widget easily without GObject,
// we assume the structure we built.
func ToContainer(w *gtk.Widget) *gtk.Container {
	// Attempt cast
	cWidget, _ := w.Cast()
	c, ok := cWidget.(*gtk.Container)
	if ok {
		return c
	}
	return nil
}

func clientsEqual(a, b []ipc.Client) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Address != b[i].Address {
			return false
		}
		if a[i].Workspace.Id != b[i].Workspace.Id {
			return false
		}
	}
	return true
}

func debugLog(format string, v ...interface{}) {
	// DISABLED for performance (Prevent UI Freeze)
	// if false {
	// 	f, err := os.OpenFile("/tmp/hypr-dock-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// 	if err == nil {
	// 		defer f.Close()
	// 		msg := fmt.Sprintf(format, v...)
	// 		timestamp := time.Now().Format("15:04:05.000")
	// 		f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
	// 	}
	// }
}
