package switcher

import (
	"sort"

	"hypr-dock/pkg/ipc"

	"github.com/gotk3/gotk3/gtk"
)

// loadData fetches clients from Hyprland and organizes them by workspace
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
		debugLog("Error loading clients: %v", err)
		return
	}

	// Filter & Group
	var filtered []ipc.Client
	s.workspaceMap = make(map[int][]int)

	// NOTE: We removed the monitor filtering logic here to ensure ALL windows
	// from ALL workspaces are loaded immediately for preview caching.
	// This eliminates the 2+ second delay when navigating between workspaces.
	// The ShowAllMonitors config can be used later for display filtering if needed.

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
			// Load ALL windows for caching - no monitor filtering
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
	logTiming("Data loaded: %d clients across %d workspaces", len(s.clients), len(s.workspaces))
	debugLog("DATA LOADED: %d Clients, %d Workspaces. Selected Index: %d. CycleWorkspaces=%v", len(s.clients), len(s.workspaces), s.selected, s.config.CycleWorkspaces)
}

// clientsEqual compares two client slices for equality
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
