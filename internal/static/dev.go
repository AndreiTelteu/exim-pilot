//go:build !embed
// +build !embed

package static

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Handler handles serving static files from filesystem in development
type Handler struct {
	root string
}

// NewHandler creates a new static file handler for development
func NewHandler() *Handler {
	// In development, serve from web/dist directory
	root := "dist"

	// Check if the directory exists, if not create a placeholder
	if _, err := os.Stat(root); os.IsNotExist(err) {
		os.MkdirAll(root, 0755)
		// Create a simple index.html for development
		indexContent := `<!DOCTYPE html>
<html>
<head>
    <title>Exim Control Panel - Development</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
        .container { max-width: 600px; margin: 0 auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Exim Control Panel</h1>
        <p>Frontend not built yet. Choose one of the following options:</p>
        <ul>
            <li>Run <code>./make build-frontend</code> to build the frontend for embedded serving</li>
            <li>Run <code>cd web && npm run dev</code> for development server on port 3000</li>
            <li>Run <code>./make build</code> to build the complete production binary</li>
        </ul>
        <p>API endpoints are available at <code>/api/v1/</code></p>
    </div>
</body>
</html>`
		os.WriteFile(filepath.Join(root, "index.html"), []byte(indexContent), 0644)
	}

	return &Handler{
		root: root,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Build the full file path
	filePath := filepath.Join(h.root, strings.TrimPrefix(upath, "/"))

	// Log what we're trying to serve for debugging
	//log.Printf("Static handler: Trying to open file: %s", strings.TrimPrefix(upath, "/"))

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// If file not found and it's not an API route, serve index.html for SPA routing
		if !strings.HasPrefix(upath, "/api/") && !strings.HasPrefix(upath, "/ws") {
			filePath = filepath.Join(h.root, "index.html")
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
		} else {
			// For API routes, return 404 immediately without trying to serve a file
			http.NotFound(w, r)
			return
		}
	}

	// Set appropriate content type and cache headers
	setContentType(w, upath)
	setCacheHeaders(w, path.Ext(upath))

	// Serve the file
	http.ServeFile(w, r, filePath)
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
