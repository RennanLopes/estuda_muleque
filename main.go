package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// variavel assets para embutir os arquivos compilados do frontend (HTML, CSS, JS)
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Cria uma instância da estrutura do aplicativo
	app := NewApp()

	// Cria o aplicativo com as opções
	err := wails.Run(&options.App{
		Title:            "Estuda Muleque!",
		Width:            1024,
		Height:           768,
		WindowStartState: options.Fullscreen,
		AlwaysOnTop:      true,
		DisableResize:    true,
		Frameless:        true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
