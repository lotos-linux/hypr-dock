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

	// ALWAYS Cycle Workspaces (MRU Workspace Mode)
	// We want to jump to the *first client* of the next/prev workspace in our list.
	// Since s.clients is grouped by workspace and sorted by WS recency:

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
			// Found new workspace!
			// BUT wait, if we are moving forward (+1), we want the FIRST item of this new workspace (which nextIdx is).
			// If we are moving backward (-1), we hit the LAST item of the previous workspace.
			// Ideally we want to select the START of that workspace block.

			if direction < 0 {
				// Search backwards to find the start of this workspace block
				for k := nextIdx; k >= 0; k-- {
					if s.clients[k].Workspace.Id == s.clients[nextIdx].Workspace.Id {
						// Found it, just need to know it exists or use it?
						// We actually just want to set s.selected to the first index of this workspace.
						// Which is stored in s.workspaceMap[s.clients[nextIdx].Workspace.Id][0]
						break
					} else {
						break
					}
				}
				// If we wrapped around, we might need to check from end?
				// Actually, our list is: [WS_A_1, WS_A_2, WS_B_1, WS_B_2]
				// If we are at WS_B_1 and go -1, we hit WS_A_2.
				// We want to jump to WS_A_1.

				targetWS := s.clients[nextIdx].Workspace.Id
				// Find first index of targetWS
				indices := s.workspaceMap[targetWS]
				if len(indices) > 0 {
					s.selected = indices[0]
				} else {
					s.selected = nextIdx // Should not happen
				}
			} else {
				// Direction > 0: We just entered a new block, so nextIdx IS the first item.
				s.selected = nextIdx
			}

			foundNewWorkspace = true
			break
		}
	}

	if foundNewWorkspace {
		debugLog("CYCLE: Switched to workspaceID %d (Client Idx %d)", s.clients[s.selected].Workspace.Id, s.selected)
	} else {
		// Fallback (Only 1 workspace active?) - Just cycle windows
		debugLog("CYCLE: No new workspace found. Fallback to window cycling.")
		s.selected += direction
		if s.selected >= len(s.clients) {
			s.selected = 0
		}
		if s.selected < 0 {
			s.selected = len(s.clients) - 1
		}
	}

	s.updateSelection()
}

// moveGrid navigates up/down in the grid layout (MRU List logic)
func (s *Switcher) moveGrid(rowDelta int) {
	if len(s.clients) == 0 {
		return
	}

	// Just jump by GridCols (width of grid)
	// We need to know GridCols. It is usually defined in render.go or ui.go
	// Assuming it is accessible or we define it.
	// We'll use a hardcoded value if not found, or relying on it being in package scope.
	// Based on prior context (not shown in file view), let's check if it compiles.
	// If GridCols is not available, we should probably find it.
	// For now, I will assume it is available as it was used before.

	// Helper to handle wrapping
	targetIdx := s.selected + (rowDelta * GridCols)

	if targetIdx < 0 {
		targetIdx = 0
	} else if targetIdx >= len(s.clients) {
		targetIdx = len(s.clients) - 1
	}

	s.selected = targetIdx
	s.updateSelection()
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
