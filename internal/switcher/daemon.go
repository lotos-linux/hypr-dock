package switcher

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

// setupSingleton checks if another instance is running and sends signal to it
// Returns true if this should be the daemon, false if signal was sent to existing daemon
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

// setupSignalHandler sets up SIGUSR1 signal handling for cycling
func (s *Switcher) setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)
	go func() {
		for range sigChan {
			glib.IdleAdd(func() {
				s.handleSignal()
			})
		}
	}()
}

// handleSignal processes SIGUSR1 signal to show/cycle the switcher
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

// checkModifiers polls keyboard modifier state to detect Alt release
// Returns true to continue polling, false to stop
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

// startPolling begins the modifier key polling loop
func (s *Switcher) startPolling() {
	glib.TimeoutAdd(50, s.checkModifiers)
}
