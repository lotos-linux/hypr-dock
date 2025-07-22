package layerinfo

import (
	"fmt"
	"hypr-dock/pkg/ipc"

	"github.com/goccy/go-json"
)

type Layer struct {
	Address   string `json:"address"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	W         int    `json:"w"`
	H         int    `json:"h"`
	Namespace string `json:"namespace"`
	Pid       int    `json:"pid"`

	Monitor string
	Layer   string
}

type Monitor struct {
	Levels map[string][]Layer `json:"levels"`
}

func GetDock() (*Layer, error) {
	return Get("hypr-dock")
}

func Get(namespace string) (*Layer, error) {
	jsonData, err := ipc.Hyprctl("j/layers")
	if err != nil {
		return nil, err
	}

	var monitors map[string]Monitor

	err = json.Unmarshal(jsonData, &monitors)
	if err != nil {
		return nil, err
	}

	for monitorName, monitor := range monitors {
		for layerName, layers := range monitor.Levels {
			for _, layer := range layers {
				if layer.Namespace == namespace {
					layer.Monitor = monitorName
					layer.Layer = layerName
					return &layer, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("%s layer not found", namespace)
}

func GetMonitor() *ipc.Monitor {
	dock, _ := GetDock()
	monitors, _ := ipc.GetMonitors()
	for _, monitor := range monitors {
		if monitor.Name == dock.Monitor {
			return &monitor
		}
	}
	return nil
}
