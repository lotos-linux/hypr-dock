# Debug Timing Guide

This document explains how to use the timing debug feature to analyze the performance of the hypr-dock switcher.

## Quick Start

### Method 1: Using the debug script (Recommended)
```bash
./debug-timing.sh
```

This will:
- Clear any old timing logs
- Run the switcher with timing enabled
- Display timing information in real-time in the terminal
- Save the complete log to `/tmp/hypr-dock-timing.log`

### Method 2: Manual execution
```bash
# Clear old log
rm -f /tmp/hypr-dock-timing.log

# Run with timing enabled
HYPR_DOCK_DEBUG_TIMING=1 hypr-dock --switcher

# View the log
cat /tmp/hypr-dock-timing.log
```

### Method 3: Watch the log in real-time
In one terminal:
```bash
tail -f /tmp/hypr-dock-timing.log
```

In another terminal:
```bash
HYPR_DOCK_DEBUG_TIMING=1 hypr-dock switcher
```

## What Gets Logged

The timing system tracks the following events with millisecond precision:

### Initialization Phase
- GTK initialization
- Switcher instance creation
- Config loading
- Singleton lock acquisition
- Signal handler setup
- Wayland connection
- Window initialization
- Main layout creation

### Data Loading Phase
- Workspace data loading
- Rendering start/complete

### UI Display Phase
- Event handlers setup
- Window display (when GUI becomes visible)
- Key polling start
- GTK main loop entry

### Async Operations (Icons)
- `[ICON]` Icon load start for each application
- `[ICON]` Icon loaded from file/theme/fallback

### Async Operations (Screenshots)
- `[SCREENSHOT]` Screenshot capture start for each window
- `[SCREENSHOT]` Screenshot captured successfully

## Understanding the Output

Each log line shows:
```
[+XXX.XXX ms] Event description
```

Where `+XXX.XXX ms` is the time elapsed since the switcher started.

### Example Output
```
[+0.000 ms] === HYPR-DOCK SWITCHER TIMING DEBUG ===
[+0.123 ms] Starting GTK initialization
[+45.678 ms] GTK initialized
[+46.234 ms] Creating Switcher instance
[+47.890 ms] Switcher instance created
...
[+150.456 ms] Window displayed - GUI is now visible
[+151.234 ms] [ICON] Starting async load for: firefox
[+152.789 ms] [SCREENSHOT] Starting async capture for: 0x12345678
[+165.432 ms] [ICON] Loaded from file: firefox
[+201.567 ms] [SCREENSHOT] Captured successfully for: 0x12345678
...
[+500.000 ms] === TOTAL TIME: 500.000 ms ===
```

## Performance Analysis Tips

1. **GUI Visibility Time**: Look for "Window displayed - GUI is now visible" to see how long until the user sees something

2. **Icon Loading**: Count `[ICON]` events to see:
   - How many icons are being loaded
   - Which icons load from file vs theme vs fallback
   - How long each icon takes

3. **Screenshot Loading**: Count `[SCREENSHOT]` events to see:
   - How many screenshots are being captured
   - Which windows fail to capture
   - How long screenshots take to appear

4. **Bottlenecks**: Large gaps between timestamps indicate slow operations

## Troubleshooting

**No timing output?**
- Make sure `HYPR_DOCK_DEBUG_TIMING=1` is set
- Check that the binary was rebuilt after adding timing code

**Log file empty?**
- The log is written in real-time, check `/tmp/hypr-dock-timing.log` while running
- Make sure you have write permissions to `/tmp/`

**Too much output?**
- The timing logs are designed to be comprehensive
- Use `grep` to filter specific events:
  ```bash
  cat /tmp/hypr-dock-timing.log | grep ICON
  cat /tmp/hypr-dock-timing.log | grep SCREENSHOT
  ```
