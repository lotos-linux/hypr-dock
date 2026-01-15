package switcher

import (
	"fmt"
	"os"
	"time"
)

var (
	// Global timing variables
	debugTimingEnabled = false
	debugStartTime     time.Time
	debugLogFile       *os.File
)

// initDebugTiming initializes the debug timing system
func initDebugTiming() {
	// Check if debug timing is enabled via environment variable
	if os.Getenv("HYPR_DOCK_DEBUG_TIMING") == "1" {
		debugTimingEnabled = true
		debugStartTime = time.Now()

		// Clear and open log file
		var err error
		debugLogFile, err = os.OpenFile("/tmp/hypr-dock-timing.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open timing log: %v\n", err)
			debugTimingEnabled = false
			return
		}

		logTiming("=== HYPR-DOCK SWITCHER TIMING DEBUG ===")
		logTiming("Start time: %s", debugStartTime.Format("15:04:05.000000"))
	}
}

// closeDebugTiming closes the debug timing log file
func closeDebugTiming() {
	if debugTimingEnabled && debugLogFile != nil {
		logTiming("=== TOTAL TIME: %.3f ms ===", time.Since(debugStartTime).Seconds()*1000)
		debugLogFile.Close()
	}
}

// logTiming logs a timing event with elapsed time since start
func logTiming(format string, v ...interface{}) {
	if !debugTimingEnabled {
		return
	}

	elapsed := time.Since(debugStartTime).Seconds() * 1000 // Convert to milliseconds
	msg := fmt.Sprintf(format, v...)
	line := fmt.Sprintf("[+%.3f ms] %s\n", elapsed, msg)

	if debugLogFile != nil {
		debugLogFile.WriteString(line)
		debugLogFile.Sync() // Flush immediately for real-time viewing
	}

	// Also print to stderr for terminal viewing
	fmt.Fprint(os.Stderr, line)
}

// debugLog writes debug messages to a log file (currently disabled for performance)
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

// enableDebugLog enables debug logging to file (for debugging purposes)
func enableDebugLog(format string, v ...interface{}) {
	f, err := os.OpenFile("/tmp/hypr-dock-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		msg := fmt.Sprintf(format, v...)
		timestamp := time.Now().Format("15:04:05.000")
		f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
	}
}
