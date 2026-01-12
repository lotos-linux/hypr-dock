package switcher

import (
	"fmt"
	"hypr-dock/pkg/ipc"
)

// getWindowFingerprint creates a unique fingerprint for a window based on its state
func getWindowFingerprint(client ipc.Client) string {
	return fmt.Sprintf("%s_%d_%dx%d",
		client.Address,
		client.Workspace.Id,
		client.Size[0],
		client.Size[1])
}
