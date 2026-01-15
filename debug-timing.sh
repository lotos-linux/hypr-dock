#!/bin/bash
# Debug timing script for hypr-dock switcher
# This script runs the switcher with timing debug enabled and shows the log in real-time

echo "=== Starting hypr-dock switcher with timing debug ==="
echo "Log file: /tmp/hypr-dock-timing.log"
echo "Press Ctrl+C to stop"
echo ""

# Clear old log
rm -f /tmp/hypr-dock-timing.log

# Run with debug timing enabled
# The timing output will appear in the terminal and in the log file
HYPR_DOCK_DEBUG_TIMING=1 /usr/bin/hypr-dock --switcher

# Show final log
echo ""
echo "=== Final timing log ==="
cat /tmp/hypr-dock-timing.log
