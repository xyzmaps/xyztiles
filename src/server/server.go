package server

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"

	"org.xyzmaps.xyztiles/src/imagery"
)

// Server represents the HTTP tile server
type Server struct {
	basemap *imagery.BaseMap
	port    int
	mux     *http.ServeMux
}

// Config holds server configuration
type Config struct {
	Port      int
	ImagePath string
}

// New creates a new tile server with the given configuration
func New(cfg Config) (*Server, error) {
	// Load the base map
	basemap, err := imagery.LoadJPEG(cfg.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base map: %w", err)
	}

	log.Printf("Loaded base map: %dx%d pixels from %s", basemap.Width(), basemap.Height(), cfg.ImagePath)

	s := &Server{
		basemap: basemap,
		port:    cfg.Port,
		mux:     http.NewServeMux(),
	}

	// Register handlers
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/tile/", s.handleTile)

	return s, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting tile server on http://localhost%s", addr)
	log.Printf("Tile endpoint: http://localhost%s/{z}/{x}/{y}.png", addr)
	return http.ListenAndServe(addr, s.mux)
}

// handleRoot serves the root endpoint (will be Leaflet viewer later)
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// Try to parse as tile request
		s.handleTileRequest(w, r, r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>xyztiles - Tile Server</title>
</head>
<body>
    <h1>xyztiles Tile Server</h1>
    <p>Server is running. Tile endpoint: <code>/{z}/{x}/{y}.png</code></p>
    <p>Base map: %dx%d pixels</p>
    <p>Example tiles:</p>
    <ul>
        <li><a href="/0/0/0.png">Zoom 0 (world)</a></li>
        <li><a href="/1/0/0.png">Zoom 1, tile 0,0</a></li>
        <li><a href="/2/1/1.png">Zoom 2, tile 1,1</a></li>
    </ul>
</body>
</html>`, s.basemap.Width(), s.basemap.Height())
}

// handleTile serves tile requests from /tile/{z}/{x}/{y}.png
func (s *Server) handleTile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tile/")
	s.handleTileRequest(w, r, "/"+path)
}

// handleTileRequest processes a tile request from a path like /{z}/{x}/{y}.png
func (s *Server) handleTileRequest(w http.ResponseWriter, r *http.Request, path string) {
	// Parse tile coordinates from path
	z, x, y, err := parseTilePath(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid tile path: %v", err), http.StatusBadRequest)
		return
	}

	// Extract the tile
	tile, err := s.basemap.ExtractTile(z, x, y)
	if err != nil {
		log.Printf("Error extracting tile %d/%d/%d: %v", z, x, y, err)
		http.Error(w, fmt.Sprintf("Failed to generate tile: %v", err), http.StatusNotFound)
		return
	}

	// Set cache headers (tiles are immutable for a given image)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hours

	// Encode as PNG
	if err := png.Encode(w, tile); err != nil {
		log.Printf("Error encoding tile %d/%d/%d: %v", z, x, y, err)
		http.Error(w, "Failed to encode tile", http.StatusInternalServerError)
		return
	}

	log.Printf("Served tile: %d/%d/%d", z, x, y)
}

// parseTilePath parses a tile path like /1/2/3.png into z, x, y coordinates
func parseTilePath(path string) (z, x, y int, err error) {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Split by /
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("expected path format /{z}/{x}/{y}.png, got %s", path)
	}

	// Parse z
	z, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid zoom level: %w", err)
	}

	// Parse x
	x, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid x coordinate: %w", err)
	}

	// Parse y (remove .png extension)
	yStr := parts[2]
	if !strings.HasSuffix(yStr, ".png") {
		return 0, 0, 0, fmt.Errorf("tile path must end with .png, got %s", yStr)
	}
	yStr = strings.TrimSuffix(yStr, ".png")

	y, err = strconv.Atoi(yStr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid y coordinate: %w", err)
	}

	return z, x, y, nil
}

// Handler returns the http.Handler for the server (useful for testing)
func (s *Server) Handler() http.Handler {
	return s.mux
}
