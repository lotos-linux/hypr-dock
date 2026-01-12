package switcher

import (
	"log"
	"os"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"hypr-dock/pkg/ipc"
	"hypr-dock/pkg/wl"

	"github.com/hashicorp/go-hclog"
)

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
	selected        int                          // Index in s.clients
	widgets         []*gtk.Widget                // Map client index to its widget
	app             *wl.App                      // Single wayland app connection
	debugBox        *gtk.EventBox                // For visual debugging of Alt key
	altWasHeld      bool                         // State for polling
	startTime       time.Time                    // Grace period
	config          Config                       // Loaded config
	screenshotCache map[string]*CachedScreenshot // Cache screenshots with timestamp

	// Daemon State
	visible       bool
	iconPathCache map[string]string
	renderGen     int
}

// CachedScreenshot stores a screenshot with its capture time
type CachedScreenshot struct {
	Pixbuf    *gdk.Pixbuf
	Timestamp time.Time
}

func Run() {
	// Initialize debug timing system
	initDebugTiming()
	defer closeDebugTiming()

	logTiming("Starting GTK initialization")
	gtk.Init(nil)
	logTiming("GTK initialized")

	logTiming("Creating Switcher instance")
	// Create Switcher instance
	s := &Switcher{
		selected:        0,
		workspaceMap:    make(map[int][]int),
		monitorMap:      make(map[int]ipc.Monitor),
		altWasHeld:      true, // ASSUME Held on start
		startTime:       time.Now(),
		config:          LoadConfig(),
		visible:         true,
		iconPathCache:   make(map[string]string),
		screenshotCache: make(map[string]*CachedScreenshot),
	}
	logTiming("Switcher instance created")

	log.Printf("------- SWITCHER CONFIG -------")
	log.Printf("Font Size: %d", s.config.FontSize)
	log.Printf("Size: %d%% x %d%%", s.config.WidthPercent, s.config.HeightPercent)
	log.Printf("Preview Width: %dpx", s.config.PreviewWidth)
	log.Printf("Show All Monitors: %v", s.config.ShowAllMonitors)
	log.Printf("-------------------------------")

	// Singleton Check
	logTiming("Checking singleton lock")
	if !setupSingleton(s) {
		return
	}
	defer os.Remove("/tmp/hypr-dock-switcher.lock")
	logTiming("Singleton lock acquired")

	// Setup signal handling for cycling
	logTiming("Setting up signal handler")
	s.setupSignalHandler()
	logTiming("Signal handler setup complete")

	// Connect to Wayland
	logTiming("Connecting to Wayland")
	app, err := wl.NewApp(hclog.Default())
	if err != nil {
		log.Printf("Failed to connect to Wayland: %v", err)
		logTiming("Wayland connection failed: %v", err)
	} else {
		s.app = app
		defer app.Close()
		logTiming("Wayland connected")
	}

	// Initialize window and UI
	logTiming("Initializing window")
	if err := s.initWindow(); err != nil {
		log.Fatal("Failed to initialize window:", err)
	}
	logTiming("Window initialized")

	// Create main layout
	logTiming("Creating main layout")
	s.createMainLayout()
	logTiming("Main layout created")

	// Load data and render
	logTiming("Loading workspace data")
	s.loadData()
	logTiming("Workspace data loaded")
	logTiming("Starting render")
	s.render()
	logTiming("Render complete (icons/screenshots will load async)")

	// Setup event handlers
	logTiming("Setting up event handlers")
	s.setupEventHandlers()
	logTiming("Event handlers setup complete")

	// Show window
	logTiming("Showing window")
	s.window.ShowAll()
	s.window.SetKeepAbove(true)
	s.window.Present()
	s.window.ShowAll()
	logTiming("Window displayed - GUI is now visible")

	// Start polling for modifier keys
	logTiming("Starting key polling")
	s.startPolling()
	logTiming("Key polling started")

	logTiming("Entering GTK main loop")
	gtk.Main()
}
