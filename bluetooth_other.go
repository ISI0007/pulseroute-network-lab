//go:build desktop && !windows

package main

import "errors"

type BluetoothCapability struct {
	Available bool   `json:"available"`
	Method    string `json:"method"`
	Detail    string `json:"detail"`
}

func (a *App) BluetoothStatus() BluetoothCapability {
	return BluetoothCapability{Method: "Platform Bluetooth sharing", Detail: "Use the operating system's supported sharing tools, or PulseRoute nearby sharing over Wi-Fi."}
}
func (a *App) OpenBluetoothTransfer() error {
	return errors.New("the integrated Bluetooth assistant is available in the Windows build only")
}
