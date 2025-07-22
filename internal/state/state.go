package state

import (
	"hypr-dock/internal/itemsctl"
	"hypr-dock/internal/pkg/timer"
	"hypr-dock/internal/pvctl"
	"hypr-dock/internal/settings"
	"sync"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type State struct {
	Settings       settings.Settings
	Window         *gtk.Window
	SignalHandlers map[string]glib.SignalHandle
	DetectArea     *gtk.Window
	ItemsBox       *gtk.Box
	Orientation    gtk.Orientation
	Edge           layershell.LayerShellEdgeFlags
	DockHideTimer  *timer.Timer
	List           *itemsctl.List
	Special        bool
	PV             *pvctl.PV
	ContextOpen    bool
	mu             sync.Mutex
}

func New(settings settings.Settings) *State {
	return &State{
		Settings:      settings,
		DockHideTimer: timer.New(),
		List:          itemsctl.New(),
		PV:            pvctl.New(settings),
	}
}

func (s *State) GetList() *itemsctl.List {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.List
}

func (s *State) GetDockHideTimer() *timer.Timer {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.DockHideTimer
}

func (s *State) AddSignalHandler(name string, id glib.SignalHandle) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.SignalHandlers == nil {
		s.SignalHandlers = make(map[string]glib.SignalHandle)
	}
	s.SignalHandlers[name] = id
}

func (s *State) RemoveSignalHandler(name string, window *gtk.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, exists := s.SignalHandlers[name]; exists {
		window.HandlerDisconnect(id)
		delete(s.SignalHandlers, name)
	}
}

func (s *State) SetEdge(edge layershell.LayerShellEdgeFlags) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Edge = edge
}

func (s *State) GetEdge() layershell.LayerShellEdgeFlags {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Edge
}

func (s *State) SetSettings(settings settings.Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Settings = settings
}

func (s *State) GetSettings() settings.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Settings
}

func (s *State) GetPinned() *[]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.Settings.PinnedApps
}

func (s *State) Update(fn func(*State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s)
}

func (s *State) SetWindow(window *gtk.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Window = window
}

func (s *State) SetDetectArea(window *gtk.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DetectArea = window
}

func (s *State) SetItemsBox(box *gtk.Box) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ItemsBox = box
}

func (s *State) SetOrientation(orientation gtk.Orientation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Orientation = orientation
}

func (s *State) SetSpecial(is bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Special = is
}

func (s *State) GetWindow() *gtk.Window {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Window
}

func (s *State) GetDetectArea() *gtk.Window {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.DetectArea
}

func (s *State) GetItemsBox() *gtk.Box {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ItemsBox
}

func (s *State) GetOrientation() gtk.Orientation {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Orientation
}

func (s *State) GetSpecial() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Special
}

func (s *State) GetPV() *pvctl.PV {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.PV
}

func (s *State) GetContextOpen() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.ContextOpen
}

func (s *State) SetContextOpen(flag bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ContextOpen = flag
}
