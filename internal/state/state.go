package state

import (
	"hypr-dock/internal/itemsctl"
	"hypr-dock/internal/layering"
	"hypr-dock/internal/pvctl"
	"hypr-dock/internal/settings"
	"sync"

	"github.com/gotk3/gotk3/gtk"
)

type State struct {
	settings settings.Settings
	window   *gtk.Window
	layerctl *layering.Control
	itemsBox *gtk.Box
	list     *itemsctl.List
	pv       *pvctl.PV
	mu       sync.Mutex
}

func New(settings settings.Settings) *State {
	return &State{
		settings: settings,
		list:     itemsctl.New(),
		pv:       pvctl.New(settings),
	}
}

func (s *State) SetLayerctl(ctl *layering.Control) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.layerctl = ctl
}

func (s *State) GetLayerctl() *layering.Control {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.layerctl
}

func (s *State) GetList() *itemsctl.List {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.list
}

func (s *State) SetSettings(settings settings.Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = settings
}

func (s *State) GetSettings() settings.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings
}

func (s *State) GetPinned() *[]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.settings.PinnedApps
}

func (s *State) Update(fn func(*State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s)
}

func (s *State) SetWindow(window *gtk.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.window = window
}

func (s *State) SetItemsBox(box *gtk.Box) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.itemsBox = box
}

func (s *State) GetWindow() *gtk.Window {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.window
}

func (s *State) GetItemsBox() *gtk.Box {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.itemsBox
}

func (s *State) GetPV() *pvctl.PV {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.pv
}
