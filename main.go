//go:build desktop

package main

import (
	"context"
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"log"
	"os"
	"path/filepath"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	exe, e := os.Executable()
	if e != nil {
		log.Fatal(e)
	}
	dir := os.Getenv("PULSEROUTE_DATA")
	if dir == "" {
		dir = filepath.Join(filepath.Dir(exe), "data")
	}
	app, e := newApp(dir)
	if e != nil {
		log.Fatal(e)
	}
	if e = wails.Run(&options.App{Title: "PulseRoute ? Network Laboratory", Width: 1380, Height: 900, MinWidth: 1000, MinHeight: 720, BackgroundColour: options.NewRGB(10, 15, 25), AssetServer: &assetserver.Options{Assets: assets}, OnStartup: app.startup, OnShutdown: func(ctx context.Context) { app.shutdown(ctx) }, Bind: []interface{}{app}, DragAndDrop: &options.DragAndDrop{EnableFileDrop: true, DisableWebViewDrop: true}, Windows: &windows.Options{WebviewUserDataPath: filepath.Join(dir, "webview")}}); e != nil {
		log.Fatal(e)
	}
}
