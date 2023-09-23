package main

import (
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	log := logger.NewFileLogger("log.txt")

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "鼠标连点器",
		Width:  350,
		Height: 330,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Logger:     log,
		OnDomReady: app.OnDomReady,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
