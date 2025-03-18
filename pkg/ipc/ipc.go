package ipc

import (
	"log"
	"net"
	"time"
)

func Hyprctl(cmd string) (response []byte, err error) {
	conn, err := net.Dial("unix", getUnixSockAdress())
	if err != nil {
		return nil, err
	}

	message := []byte(cmd)
	_, err = conn.Write(message)
	if err != nil {
		return nil, err
	}

	response = make([]byte, 102400)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	return response[:n], nil
}

func InitHyprEvents() {
	for {
		unixConnect, err := net.DialUnix("unix", nil, getUnixSock2Adress())
		if err != nil {
			log.Printf("Failed to connect to Unix socket: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer unixConnect.Close()

		for {
			buffer := make([]byte, 10240)
			unixNumber, err := unixConnect.Read(buffer)
			if err != nil {
				log.Printf("Error reading from Unix socket: %v. Reconnecting...", err)
				break
			}

			events := splitEvent(string(buffer[:unixNumber]))
			for _, event := range events {
				go DispatchEvent(event)
			}
		}
	}
}
