//go:build !embed
// +build !embed

package web

import (
	"log"
)

// InitEmbeddedAssets is a no-op for development builds
func InitEmbeddedAssets() {
	log.Printf("Development build: Skipping embedded assets initialization")
}
