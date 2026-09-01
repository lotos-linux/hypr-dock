package switcher

import "hypr-dock/internal/desktop"

// resolveIconName returns the icon declared by the application's desktop entry
// when available. The window class remains the fallback for applications that
// cannot be resolved through desktop metadata.
func resolveIconName(className string) string {
	app, err := desktop.New(className)
	if err != nil {
		return className
	}

	if iconName := app.GetIcon(); iconName != "" {
		return iconName
	}

	return className
}
