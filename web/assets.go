//go:build embed
// +build embed

package web

import (
	"embed"
	"io/fs"
	"log"

	"github.com/andreitelteu/exim-pilot/internal/static"
)

// StaticAssets holds the embedded static files from the web/dist directory
//
//go:embed dist
var StaticAssets embed.FS

// InitEmbeddedAssets initializes the embedded static assets for production builds
func InitEmbeddedAssets() {
	log.Printf("Initializing embedded assets...")
	// Initialize embedded static assets for production builds
	distFS, err := fs.Sub(StaticAssets, "dist")
	if err != nil {
		log.Printf("ERROR: Failed to initialize embedded assets: %v", err)
		return
	}

	log.Printf("SUCCESS: Embedded assets initialized, setting filesystem...")
	// Set the embedded filesystem in the static package
	static.SetEmbeddedFS(distFS)
	log.Printf("SUCCESS: Embedded filesystem set in static package")
}
