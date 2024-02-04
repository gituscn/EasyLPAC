package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"runtime"
)

const StatusProcess = 1
const StatusReady = 0

var SelectedProfile int
var SelectedNotification int

var RefreshNeeded bool
var ProfileMaskNeeded bool
var NotificationMaskNeeded bool
var ProfileStateAllowDisable bool

var StatusChan chan int
var LockButtonChan chan bool

func RefreshProfile() {
	var err error
	Profiles, err = LpacProfileList()
	if err != nil {
		ShowErrDialog(err)
	}
	// 刷新 List
	ProfileList.Refresh()
	ProfileList.UnselectAll()
}

func RefreshNotification() {
	var err error
	Notifications, err = LpacNotificationList()
	if err != nil {
		ShowErrDialog(err)
	}
	// 刷新 List
	NotificationList.Refresh()
	NotificationList.UnselectAll()
}

func RefreshChipInfo() {
	var err error
	ChipInfo, err = LpacChipInfo()
	if err != nil {
		ShowErrDialog(err)
	}

	convertToString := func(value interface{}) string {
		if value == nil {
			return "<not set>"
		}
		if str, ok := value.(string); ok {
			return str
		}
		return "<not set>"
	}

	EidLabel.SetText(fmt.Sprintf("EID: %s", ChipInfo.EidValue))
	DefaultDpAddressLabel.SetText(fmt.Sprintf("Default SM-DP+ Address:  %s", convertToString(ChipInfo.EuiccConfiguredAddresses.DefaultDpAddress)))
	RootDsAddressLabel.SetText(fmt.Sprintf("Root SM-DS Address:  %s", convertToString(ChipInfo.EuiccConfiguredAddresses.RootDsAddress)))
	bytes, err := json.MarshalIndent(ChipInfo.EUICCInfo2, "", "  ")
	if err != nil {
		ShowErrDialog(fmt.Errorf("chip Info: failed to decode EUICCInfo2\n%s", err))
	}
	CopyEidButton.Show()
	SetDefaultSmdpButton.Show()
	EuiccInfo2Entry.Show()
	ViewCertInfoButton.Show()

	// EuiccInfo2TextGrid.SetText(string(bytes))
	EuiccInfo2Entry.SetText(string(bytes))
	// 计算剩余空间
	freeSpace := float64(ChipInfo.EUICCInfo2.ExtCardResource.FreeNonVolatileMemory) / 1024
	FreeSpaceLabel.SetText(fmt.Sprintf("Free space: %.2f KB", math.Round(freeSpace*100)/100))
}

func RefreshApduDriver() {
	var err error
	ApduDrivers, err = LpacDriverApduList()
	if err != nil {
		ShowErrDialog(err)
	}
	var options []string
	for _, d := range ApduDrivers {
		options = append(options, d.Name)
	}
	ApduDriverSelect.SetOptions(options)
	ApduDriverSelect.ClearSelected()
	ConfigInstance.DriverIFID = ""
	ApduDriverSelect.Refresh()
}

func OpenLog() {
	var err error

	switch runtime.GOOS {
	case "windows":
		err = exec.Command("explorer", ConfigInstance.LogDir).Start()
	case "darwin":
		err = exec.Command("open", ConfigInstance.LogDir).Start()
	case "linux":
		err = exec.Command("xdg-open", ConfigInstance.LogDir).Start()
	default:
		err = fmt.Errorf("unsupported platform, please open log file manually")
		ShowErrDialog(err)
	}

	if err != nil {
		ShowErrDialog(err)
	}
}

func Refresh() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	RefreshProfile()
	RefreshNotification()
	RefreshChipInfo()
	RefreshNeeded = false
}

func UpdateStatusBar() {
	for {
		status := <-StatusChan
		switch status {
		case StatusProcess:
			StatusLabel.SetText("Processing...")
			StatusProcessBar.Start()
			StatusProcessBar.Show()
			continue
		case StatusReady:
			StatusLabel.SetText("Ready.")
			StatusProcessBar.Stop()
			StatusProcessBar.Hide()
			continue
		}
	}
}

func LockButton() {
	for {
		lock := <-LockButtonChan
		if lock {
			DownloadButton.Disable()
			DiscoveryButton.Disable()
			SetNicknameButton.Disable()
			RefreshButton.Disable()
			SwitchStateButton.Disable()
			DeleteButton.Disable()
			ProcessNotificationButton.Disable()
			RemoveNotificationButton.Disable()
			SetDefaultSmdpButton.Disable()
			ProfileMaskCheck.Disable()
			NotificationMaskCheck.Disable()
			ApduDriverSelect.Disable()
			ApduDriverRefreshButton.Disable()
		} else {
			DownloadButton.Enable()
			DiscoveryButton.Enable()
			SetNicknameButton.Enable()
			RefreshButton.Enable()
			SwitchStateButton.Enable()
			DeleteButton.Enable()
			ProcessNotificationButton.Enable()
			RemoveNotificationButton.Enable()
			SetDefaultSmdpButton.Enable()
			ProfileMaskCheck.Enable()
			NotificationMaskCheck.Enable()
			ApduDriverSelect.Enable()
			ApduDriverRefreshButton.Enable()
		}
	}
}

func SetDriverIfid(name string) {
	for _, d := range ApduDrivers {
		if name == d.Name {
			if ConfigInstance.DriverIFID == "" {
				ConfigInstance.DriverIFID = d.Env
			} else if ConfigInstance.DriverIFID == d.Env {
				// 未改变读卡器，不刷新状态
				return
			} else {
				ConfigInstance.DriverIFID = d.Env
				RefreshNeeded = true
			}
		}
	}
}
