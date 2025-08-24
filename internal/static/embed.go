//go:build embed
// +build embed

package static

import (
	"log"
	"net/http"
	"path"
	"strings"
)

// Handler handles serving embedded static files
type Handler struct {
	fileSystem http.FileSystem
}

// NewHandler creates a new static file handler
func NewHandler() *Handler {
	// Get the embedded filesystem starting from web/dist
	distFS, err := GetDistFS()
	if err != nil || distFS == nil {
		log.Printf("ERROR: Failed to get embedded filesystem: %v", err)
		// Return a handler that always returns 404 if no embedded assets
		return &Handler{
			fileSystem: nil,
		}
	}

	log.Printf("SUCCESS: Embedded filesystem initialized")
	return &Handler{
		fileSystem: http.FS(distFS),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Static handler: Request for %s", r.URL.Path)

	// If no embedded filesystem is available, return 404
	if h.fileSystem == nil {
		log.Printf("Static handler: No embedded filesystem available")
		http.NotFound(w, r)
		return
	}

	// Clean the path
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	upath = path.Clean(upath)

	// Remove leading slash for filesystem access
	if upath == "/" {
		upath = "/index.html"
	}

	filePath := strings.TrimPrefix(upath, "/")
	log.Printf("Static handler: Trying to open file: %s", filePath)

	// Try to serve the requested file
	file, err := h.fileSystem.Open(filePath)
	if err != nil {
		log.Printf("Static handler: Failed to open file %s: %v", filePath, err)
		// If file not found and it's not an API route, serve index.html for SPA routing
		if !strings.HasPrefix(upath, "/api/") && !strings.HasPrefix(upath, "/ws") {
			log.Printf("Static handler: Trying fallback to index.html")
			file, err = h.fileSystem.Open("index.html")
			if err != nil {
				log.Printf("Static handler: Failed to open index.html: %v", err)
				http.NotFound(w, r)
				return
			}
		} else {
			http.NotFound(w, r)
			return
		}
	}
	defer file.Close()

	// Get file info for proper headers
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set appropriate content type based on file extension
	setContentType(w, upath)

	// Set cache headers for static assets (not for HTML files)
	setCacheHeaders(w, path.Ext(upath))

	// Serve the file
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

// setContentType sets the appropriate content type based on file extension
func setContentType(w http.ResponseWriter, upath string) {
	ext := path.Ext(upath)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".woff":
		w.Header().Set("Content-Type", "font/woff")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	}
}

// setCacheHeaders sets appropriate cache headers
func setCacheHeaders(w http.ResponseWriter, ext string) {
	if ext != ".html" {
		// Cache static assets for 1 year
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("Expires", "Thu, 31 Dec 2037 23:55:55 GMT")
	} else {
		// No cache for HTML files to ensure SPA routing works
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
}
