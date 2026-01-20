package main

import (
	"github.com/ebiyu/wslusb/usbipd"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UIState struct {
	AttachedItems *[]usbipd.Device
	DetachedItems *[]usbipd.Device
}

func main() {
	app := tview.NewApplication()

	detachedPane := tview.NewFlex()
	detachedPane.SetTitle("Detached (Attached to Host)").SetBorder(true)

	detachedTable := tview.NewTable()
	detachedPane.AddItem(detachedTable, 0, 1, true)

	attachedPane := tview.NewFlex()
	attachedPane.SetTitle("Attached to WSL").SetBorder(true)

	attachedTable := tview.NewTable()
	attachedPane.AddItem(attachedTable, 0, 1, true)

	flex := tview.NewFlex().
		AddItem(detachedPane, 0, 1, true).
		AddItem(attachedPane, 0, 1, true)

	flex.SetBorder(true)
	flex.SetTitle("USBIPD")

	uiState := UIState{
		AttachedItems: &[]usbipd.Device{},
		DetachedItems: &[]usbipd.Device{},
	}

	updateDeviceList := func() {
		attachedTable.Clear()
		detachedTable.Clear()

		items, err := usbipd.GetDevices()
		if err != nil {
			panic(err)
		}

		attachedItems := []usbipd.Device{}
		detachedItems := []usbipd.Device{}
		for _, v := range items {
			if v.BusID == "" {
				// Not connected now
				continue
			}
			if v.Status == usbipd.Attached {
				attachedItems = append(attachedItems, v)
			} else {
				detachedItems = append(detachedItems, v)
			}
		}

		detachedTable.SetCell(0, 0, tview.NewTableCell("BusID").SetSelectable(false))
		detachedTable.SetCell(0, 1, tview.NewTableCell("Status").SetSelectable(false))
		detachedTable.SetCell(0, 2, tview.NewTableCell("DeviceID").SetSelectable(false))
		detachedTable.SetCell(0, 3, tview.NewTableCell("DeviceName").SetSelectable(false))
		for i, v := range detachedItems {
			statusText := "Not Shared"
			if v.Status == usbipd.Shared {
				statusText = "Shared"
			}
			detachedTable.SetCell(i+1, 0, tview.NewTableCell(v.BusID))
			detachedTable.SetCell(i+1, 1, tview.NewTableCell(statusText))
			detachedTable.SetCell(i+1, 2, tview.NewTableCell(v.DeviceID))
			detachedTable.SetCell(i+1, 3, tview.NewTableCell(v.DeviceName))
		}

		attachedTable.SetCell(0, 0, tview.NewTableCell("BusID").SetSelectable(false))
		attachedTable.SetCell(0, 1, tview.NewTableCell("DeviceID").SetSelectable(false))
		attachedTable.SetCell(0, 2, tview.NewTableCell("DeviceName").SetSelectable(false))

		for i, v := range attachedItems {
			attachedTable.SetCell(i+1, 0, tview.NewTableCell(v.BusID))
			attachedTable.SetCell(i+1, 1, tview.NewTableCell(v.DeviceID))
			attachedTable.SetCell(i+1, 2, tview.NewTableCell(v.DeviceName))
		}

		uiState.AttachedItems = &attachedItems
		uiState.DetachedItems = &detachedItems
	}

	updateDeviceList()

	detachedTable.SetSelectable(true, false)
	detachedTable.Select(1, 0)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'r' {
			updateDeviceList()
			return nil
		}

		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			app.Stop()
			return nil
		}

		return event
	})

	detachedTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRight || (event.Key() == tcell.KeyRune && event.Rune() == 'l') {
			attachedTable.SetSelectable(true, false)
			detachedTable.SetSelectable(false, false)
			app.SetFocus(attachedTable)
			attachedTable.Select(1, 0)
			return nil
		}

		return event
	})

	attachedTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyLeft || (event.Key() == tcell.KeyRune && event.Rune() == 'h') {
			attachedTable.SetSelectable(false, false)
			detachedTable.SetSelectable(true, false)
			app.SetFocus(detachedTable)
			detachedTable.Select(1, 0)
			return nil
		}

		return event
	})

	detachedTable.SetSelectedFunc(func(row, column int) {
		device := (*uiState.DetachedItems)[row-1]
		if device.Status == usbipd.NotShared {
			err := usbipd.BindDevice(device.BusID)
			if err != nil {
				panic(err)
			}
		}

		err := usbipd.AttachDevice(device.BusID)
		if err != nil {
			panic(err)
		}

		updateDeviceList()
	})

	attachedTable.SetSelectedFunc(func(row, column int) {
		device := (*uiState.AttachedItems)[row-1]
		err := usbipd.DetachDevice(device.BusID)
		if err != nil {
			panic(err)
		}

		updateDeviceList()
	})

	app.SetFocus(detachedTable)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

}
