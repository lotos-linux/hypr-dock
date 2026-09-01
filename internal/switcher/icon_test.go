package switcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveIconNameUsesDesktopEntryIcon(t *testing.T) {
	dataHome := t.TempDir()
	applicationsDir := filepath.Join(dataHome, "applications")
	if err := os.MkdirAll(applicationsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	desktopFile := `[Desktop Entry]
Name=Renamed App
Exec=renamed-app
Icon=renamed-app-icon
StartupWMClass=com.example.SpecialEditor
`
	if err := os.WriteFile(filepath.Join(applicationsDir, "renamed-app.desktop"), []byte(desktopFile), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", dataHome)
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("XDG_DATA_DIRS", "")

	if got := resolveIconName("com.example.SpecialEditor"); got != "renamed-app-icon" {
		t.Fatalf("resolveIconName() = %q, want %q", got, "renamed-app-icon")
	}

	if got := resolveIconName("missing.example.App"); got != "missing.example.App" {
		t.Fatalf("resolveIconName() fallback = %q, want window class", got)
	}
}
