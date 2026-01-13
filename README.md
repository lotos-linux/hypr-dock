# hypr-dock & hypr-alttab

### Essential Desktop Tools for Hyprland

This project provides two powerful tools to enhance your Hyprland experience:
1.  **hypr-dock**: An interactive, customizable dock panel.
2.  **hypr-alttab**: A Windows-style Alt-Tab switcher with workspace previews.

Translations: [`Русский`](README_RU.md)

<img width="1360" height="768" alt="Screenshot 1" src="https://github.com/user-attachments/assets/041d2cf6-13ba-4c89-a960-1903073ff2d4" />

[![YouTube](https://img.shields.io/badge/YouTube-Video-FF0000?logo=youtube)](https://youtu.be/HHUZWHfNAl0?si=ZrRv2ggnPBEBS5oY)

---

## 🚀 Easy Installation

We provide an interactive installer to help you set up one or both tools quickly.

### Prerequisites
- `go` (golang)
- `gtk3`, `gtk-layer-shell`

### Install Steps
1.  Clone the repository:
    ```bash
    git clone https://github.com/lotos-linux/hypr-dock.git
    cd hypr-dock
    ```
2.  Get dependencies:
    ```bash
    make get
    ```
3.  **Run the Installer**:
    ```bash
    ./install.sh
    ```
    Follow the prompts to install **hypr-dock**, **hypr-alttab**, or both.

---

## 🖥️ hypr-dock (The Dock)

An interactive dock panel similar to macOS or Deepin.

### Launching
Add this to your `hyprland.conf`:
```text
exec-once = hypr-dock
```
Or launch manually: `hypr-dock`

### Configuration
Config file: `~/.config/hypr-dock/config.jsonc`

Key options:
- `"Position"`: "bottom", "top", "left", "right"
- `"Layer"`: "auto" (smart hide), "exclusive-bottom" (always visible), etc.
- `"Preview"`: "static" (window screenshot) or "live" (experimental streaming).
- `"Pinned"`: List of pinned apps (edit `pinned.json` manually if needed).

### Themes
Themes are in `~/.config/hypr-dock/themes/`. Change the current theme in `config.jsonc`.

---

## 🔄 hypr-alttab (The Switcher)

A visual Alt-Tab switcher that groups windows by **Workspace** and sorts them by **Recency (MRU)**.

**Demo:** [Watch on YouTube](https://www.youtube.com/watch?v=rU1Ex1y95Rs)

### Features
- **MRU Workspace Switching**: Alt-Tab cycles through workspaces based on when you last used them.
- **Workspace Cards**: Visual groups showing all windows in a workspace.
- **Live Previews**: Fast, cached previews of your windows.

### Launching
Add this to your `hyprland.conf`:
```text
# Bind Alt+Tab to hypr-alttab
bind = ALT, Tab, exec, hypr-alttab
```

### Configuration
Config file: `~/.config/hypr-dock/switcher.jsonc`

| Parameter | Default | Description |
| :--- | :--- | :--- |
| `widthPercent` | `100` | Width of the overlay (screen percentage) |
| `heightPercent`| `60` | Height of the overlay |
| `previewWidth` | `400` | Width of individual workspace cards (px) |
| `cycleWorkspaces`| `true` | Cycle strictly between workspace cards |

---

## 🛠️ Building from Source (Advanced)

If you prefer `make` commands over the installer script:

**Build:**
```bash
make build
# Creates bin/hypr-dock and bin/hypr-alttab
```

**Install:**
```bash
make install-dock   # Install only dock
make install-alttab # Install only switcher
make install-all    # Install both
```

**Uninstall:**
```bash
make uninstall
```

## 🔐 Permissions
For window previews to work, you need to allow screencopy permissions in `hyprland.conf`:
```text
permission = /usr/bin/hypr-dock, screencopy, allow
permission = /usr/bin/hypr-alttab, screencopy, allow
``` 
See [Hyprland Permissions Wiki](https://wiki.hypr.land/Configuring/Permissions/) for details.

## ❓ Troubleshooting

### Protocol Errors / Build Issues
If you encounter errors related to Wayland protocols during the build, you may need to regenerate the Go bindings:

```bash
# Update Go Wayland library
go get -u github.com/pdf/go-wayland@latest
go mod tidy

# Install/Update the scanner tool
go install github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner@latest

# Download latest protocol XML
wget https://raw.githubusercontent.com/hyprwm/hyprland-protocols/main/protocols/hyprland-toplevel-export-v1.xml

# Generate Go code
$(go env GOPATH)/bin/go-wayland-scanner -i hyprland-toplevel-export-v1.xml -o pkg/wl/hyprland_toplevel_export.go -pkg wl
```
