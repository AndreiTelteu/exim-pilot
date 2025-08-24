package static

import (
	"io/fs"
)

var embeddedFS fs.FS

// SetEmbeddedFS sets the embedded filesystem for production builds
func SetEmbeddedFS(fsys fs.FS) {
	embeddedFS = fsys
}

// GetDistFS returns the embedded dist filesystem
func GetDistFS() (fs.FS, error) {
	if embeddedFS != nil {
		return embeddedFS, nil
	}
	return nil, fs.ErrNotExist
}
