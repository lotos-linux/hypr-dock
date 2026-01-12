package switcher

import (
	"fmt"

	"hypr-dock/pkg/ipc"
)

// cycle moves selection by direction (1 for forward, -1 for backward)
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

// moveGrid navigates up/down in the grid layout
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

// confirm switches to the selected window and hides the switcher
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
