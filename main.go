package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	if os.Getenv("WEBVIEW2_RELEASE_PATH") == "" {
		paths := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "EdgeWebView", "Application"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "EdgeWebView", "Application"),
		}
		for _, p := range paths {
			entries, err := os.ReadDir(p)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					os.Setenv("WEBVIEW2_RELEASE_PATH", filepath.Join(p, e.Name()))
					break
				}
			}
			if os.Getenv("WEBVIEW2_RELEASE_PATH") != "" {
				break
			}
		}
	}
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "vibe-dashboard",
		Width:  1280,
		Height: 860,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 20, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
		},
		Linux: &linux.Options{},
	})

	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}
}
