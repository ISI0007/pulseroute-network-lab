//go:build desktop && windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
)

type BluetoothCapability struct {
	Available bool   `json:"available"`
	Method    string `json:"method"`
	Detail    string `json:"detail"`
}

func bluetoothHelper() (string, error) {
	directory, err := windows.GetSystemDirectory()
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "fsquirt.exe")
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("Windows Bluetooth File Transfer is unavailable on this device")
	}
	return path, nil
}
func (a *App) BluetoothStatus() BluetoothCapability {
	_, err := bluetoothHelper()
	return BluetoothCapability{Available: err == nil, Method: "Windows Bluetooth File Transfer", Detail: "The Windows assistant handles pairing, device selection and the transfer. Radio availability and completion are not tracked by PulseRoute. iPhone/iPad should use Wi-Fi nearby sharing."}
}
func (a *App) OpenBluetoothTransfer() error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	path, err := bluetoothHelper()
	if err != nil {
		return err
	}
	// Fixed protected-system executable, no shell or remotely supplied arguments.
	command := exec.Command(path)
	if err = command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
