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

	// Find focused monitor
	focusedMonitorID := -1
	for _, m := range s.monitors {
		if m.Focused {
			focusedMonitorID = m.Id
			break
		}
	}

	// Filter & Group
	s.workspaceMap = make(map[int][]int)

	// Group clients by workspace temporarily to calculate recency
	tempWsMap := make(map[int][]ipc.Client)
	minFocusMap := make(map[int]int) // Workspace ID -> Minimum FocusHistoryID (Lower = More Recent)

	for _, c := range all {
		// Filter by Monitor if needed
		if !s.config.ShowAllMonitors && focusedMonitorID != -1 {
			if c.Monitor != focusedMonitorID {
				continue
			}
		}

		if c.Mapped && c.Workspace.Id > 0 {
			tempWsMap[c.Workspace.Id] = append(tempWsMap[c.Workspace.Id], c)

			// Update minimum focus ID for this workspace
			currMin, exists := minFocusMap[c.Workspace.Id]
			if !exists || c.FocusHistoryID < currMin {
				minFocusMap[c.Workspace.Id] = c.FocusHistoryID
			}
		}
	}

	// Get list of workspaces
	var ws []int
	for k := range tempWsMap {
		ws = append(ws, k)
	}

	// Sort Workspaces by Recency (Min FocusHistoryID)
	// Lower MinFocusID = Workspace was active more recently
	sort.Slice(ws, func(i, j int) bool {
		minI := minFocusMap[ws[i]]
		minJ := minFocusMap[ws[j]]
		return minI < minJ
	})
	s.workspaces = ws

	// Rebuild s.clients in the order of Sorted Workspaces
	var orderedClients []ipc.Client

	for _, wsID := range s.workspaces {
		clientsInWs := tempWsMap[wsID]

		// Sort clients WITHIN workspace by FocusHistoryID (MRU within workspace)
		sort.Slice(clientsInWs, func(i, j int) bool {
			return clientsInWs[i].FocusHistoryID < clientsInWs[j].FocusHistoryID
		})

		// Append to main list and build map
		for _, c := range clientsInWs {
			orderedClients = append(orderedClients, c)
			idx := len(orderedClients) - 1
			s.workspaceMap[wsID] = append(s.workspaceMap[wsID], idx)
		}
	}
	s.clients = orderedClients
	s.widgets = make([]*gtk.Widget, len(s.clients))

	// Selection Logic: Select the first client of the SECOND workspace (Index 1)
	// Index 0 in s.workspaces is the current active workspace.
	// Index 1 is the previous workspace.
	targetIdx := 0

	if len(s.workspaces) > 1 {
		// Select the first window of the second workspace
		targetWsID := s.workspaces[1]
		indices := s.workspaceMap[targetWsID]
		if len(indices) > 0 {
			targetIdx = indices[0]
		}
	} else if len(s.workspaces) == 1 {
		// Only one workspace, maybe multiple windows?
		// Select the second window (Index 1) if available (Previous App in same WS)
		if len(s.clients) > 1 {
			targetIdx = 1
		}
	}

	s.selected = targetIdx
	debugLog("DATA LOADED: %d Clients, %d Workspaces (MRU). Selected Index: %d.", len(s.clients), len(s.workspaces), s.selected)
	logTiming("Data loaded: Sorted by Workspace Recency")
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
