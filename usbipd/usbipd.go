package usbipd

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/ebiyu/wslusb/elevate"
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

type usbipdStatusOutput struct {
	Devices []usbipdDevice `json:"Devices"`
}

type usbipdDevice struct {
	BusId           *string `json:"BusId"`
	ClientIPAddress *string `json:"ClientIPAddress"`
	Description     string  `json:"Description"`
	InstanceId      string  `json:"InstanceId"`
	IsForced        bool    `json:"IsForced"`
	PersistedGuid   *string `json:"PersistedGuid"`
	StubInstanceId  *string `json:"StubInstanceId"`
}

func (d *usbipdDevice) GetStatus() int {
	if d.ClientIPAddress != nil {
		return Attached
	}
	if d.PersistedGuid != nil || d.IsForced {
		return Shared
	}
	return NotShared
}

func GetDevices() ([]Device, error) {
	cmd := exec.Command("usbipd.exe", "state")
	output, err := cmd.Output()
	if err != nil {
		return []Device{}, err
	}

	var statusOutput usbipdStatusOutput
	if err := json.Unmarshal(output, &statusOutput); err != nil {
		return []Device{}, err
	}

	items := []Device{}
	for _, d := range statusOutput.Devices {
		busId := ""
		if d.BusId != nil {
			busId = *d.BusId
		}
		items = append(items, Device{
			BusID:      busId,
			DeviceID:   d.InstanceId,
			DeviceName: d.Description,
			Status:     d.GetStatus(),
		})
	}

	return items, nil
}

func BindDevice(busid string) error {
	if elevate.IsElevated() {
		// Already elevated, execute directly
		cmd := exec.Command("usbipd.exe", "bind", "--busid", busid)
		_, err := cmd.Output()
		return err
	}

	// Not elevated, use UAC
	args := fmt.Sprintf("bind --busid %s", busid)
	exitCode, err := elevate.RunAsAdminWait("usbipd.exe", args)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("usbipd.exe exited with code %d", exitCode)
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
