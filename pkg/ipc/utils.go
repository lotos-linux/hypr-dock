package ipc

import (
	"errors"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func SearchClientByAddress(address string) (Client, error) {
	clients, err := GetClients()
	if err != nil {
		log.Println(err)
		return Client{}, err
	}

	for _, ipcClient := range clients {
		if ipcClient.Address == address {
			return ipcClient, nil
		}
	}

	err = errors.New("Client non found by address: " + address)
	return Client{}, err
}

func SearchMonitorByName(name string) (*Monitor, error) {
	monitors, err := GetMonitors()
	if err != nil {
		return nil, err
	}

	for _, monitor := range monitors {
		if monitor.Name == name {
			return &monitor, nil
		}
	}
	return nil, errors.New("Monitor not found by name: " + name)
}

func getHyprPathes() (XDGRuntimeDirHypr string, HIS string) {
	XDGRuntimeDirHypr = filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), "hypr")
	HIS = os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	return XDGRuntimeDirHypr, HIS
}

func getUnixSockAdress() (unixSockAdress string) {
	XDGRuntimeDirHypr, HIS := getHyprPathes()

	return filepath.Join(XDGRuntimeDirHypr, HIS, ".socket.sock")
}

func getUnixSock2Adress() (unixSock2Adress *net.UnixAddr) {
	XDGRuntimeDirHypr, HIS := getHyprPathes()

	return &net.UnixAddr{
		Name: filepath.Join(XDGRuntimeDirHypr, HIS, ".socket2.sock"),
		Net:  "unix",
	}
}

func splitEvent(multiLineEvent string) []string {
	events := strings.Split(multiLineEvent, "\n")

	var filteredEvents []string
	for _, event := range events {
		event = strings.TrimSpace(event)
		if event != "" {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents
}
