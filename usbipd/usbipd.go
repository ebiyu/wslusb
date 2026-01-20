package usbipd

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	NotShared = iota
	Shared
	Attached
)

type Device struct {
	BusID      string
	DeviceID   string
	DeviceName string
	Status     int
}

func GetDevices() ([]Device, error) {
	cmd := exec.Command("usbipd.exe", "list")
	output, err := cmd.Output()
	if err != nil {
		return []Device{}, err
	}

	// make output list
	strOutput := string(output)
	strOutputList := strings.Split(strOutput, "\n")
	for i, v := range strOutputList {
		strOutputList[i] = strings.TrimSpace(v)
	}

	// Parse the output
	begnRow := -1
	endRow := -1

	for i, v := range strOutputList {
		if v == "Connected:" {
			begnRow = i + 2
		}
		if v == "Persisted:" {
			endRow = i
		}
	}
	if begnRow == -1 || endRow == -1 {
		return []Device{}, fmt.Errorf("Could not find the beginning or end of the device list")
	}

	items := []Device{}
	for _, v := range strOutputList[begnRow:endRow] {
		cols := strings.Fields(v)
		if len(cols) < 3 {
			continue
		}
		busid, device, remainder := cols[0], cols[1], cols[2:]
		status := NotShared
		if remainder[len(remainder)-1] == "Shared" {
			status = Shared
		} else if remainder[len(remainder)-1] == "Attached" {
			status = Attached
		}
		deviceName := strings.Join(remainder[:len(remainder)-1], " ")
		items = append(items, Device{
			BusID:      busid,
			DeviceID:   device,
			DeviceName: deviceName,
			Status:     status,
		})
	}

	return items, nil
}

func BindDevice(busid string) error {
	cmd := exec.Command("usbipd.exe", "bind", "--busid", busid)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func AttachDevice(busid string) error {
	cmd := exec.Command("usbipd.exe", "attach", "--wsl", "--busid", busid)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func DetachDevice(busid string) error {
	cmd := exec.Command("usbipd.exe", "detach", "--busid", busid)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
