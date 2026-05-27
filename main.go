package main

import (
	"embed"
	"webhelper/global"
	"webhelper/service"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed lang/*
var langFS embed.FS

func main() {
	global.LangFS = langFS
	global.Init()

	appName := global.GetProcessName()
	App := service.NewApp()

	err := wails.Run(&options.App{
		Title:     appName,
		Width:     1024,
		Height:    768,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        App.Startup,
		Frameless:        true,
		Bind: []interface{}{
			App,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
