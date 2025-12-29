package server

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testImagePath = "../../res/world.topo.200407.3x5400x2700.jpg"

func TestNew(t *testing.T) {
	// Check if test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image not found at %s, skipping test", testImagePath)
		return
	}

	cfg := Config{
		Port:      8080,
		ImagePath: testImagePath,
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if srv == nil {
		t.Fatal("New() returned nil server")
	}

	if srv.port != 8080 {
		t.Errorf("Expected port 8080, got %d", srv.port)
	}

	if srv.basemap == nil {
		t.Fatal("Server basemap is nil")
	}
}

func TestNew_InvalidImage(t *testing.T) {
	cfg := Config{
		Port:      8080,
		ImagePath: "/nonexistent/image.jpg",
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error for invalid image path, got nil")
	}
}

func TestNew_EmbeddedData(t *testing.T) {
	// Check if test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image not found at %s, skipping test", testImagePath)
		return
	}

	// Read the test image into bytes (simulating embedded data)
	data, err := os.ReadFile(testImagePath)
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	cfg := Config{
		Port:         8080,
		EmbeddedData: data,
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() with embedded data failed: %v", err)
	}

	if srv == nil {
		t.Fatal("New() returned nil server")
	}

	if srv.basemap == nil {
		t.Fatal("Server basemap is nil")
	}

	t.Logf("Successfully created server with embedded data (%d bytes)", len(data))
}

func TestNew_EmbeddedDataInvalidJPEG(t *testing.T) {
	invalidData := []byte("This is not a JPEG image")

	cfg := Config{
		Port:         8080,
		EmbeddedData: invalidData,
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error for invalid embedded data, got nil")
	}
}

func TestParseTilePath(t *testing.T) {
	tests := []struct {
		path        string
		expectZ     int
		expectX     int
		expectY     int
		expectError bool
		name        string
	}{
		{"/0/0/0.png", 0, 0, 0, false, "zoom 0"},
		{"/1/0/0.png", 1, 0, 0, false, "zoom 1"},
		{"/5/10/15.png", 5, 10, 15, false, "zoom 5"},
		{"/12/2048/1024.png", 12, 2048, 1024, false, "zoom 12"},
		{"0/0/0.png", 0, 0, 0, false, "no leading slash"},

		// Error cases
		{"/0/0.png", 0, 0, 0, true, "missing coordinate"},
		{"/0/0/0/0.png", 0, 0, 0, true, "too many parts"},
		{"/a/0/0.png", 0, 0, 0, true, "invalid z"},
		{"/0/b/0.png", 0, 0, 0, true, "invalid x"},
		{"/0/0/c.png", 0, 0, 0, true, "invalid y"},
		{"/0/0/0.jpg", 0, 0, 0, true, "wrong extension"},
		{"/0/0/0", 0, 0, 0, true, "no extension"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z, x, y, err := parseTilePath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for path %s, got nil", tt.path)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if z != tt.expectZ || x != tt.expectX || y != tt.expectY {
				t.Errorf("parseTilePath(%s) = (%d, %d, %d), expected (%d, %d, %d)",
					tt.path, z, x, y, tt.expectZ, tt.expectX, tt.expectY)
			}
		})
	}
}

func TestHandleRoot(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.handleRoot(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}

	// Check Cache-Control header
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("Expected Cache-Control 'public, max-age=3600', got %s", cacheControl)
	}

	// Check that response contains expected text
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Response body is empty")
	}

	// Check for Leaflet viewer content
	expectedStrings := []string{
		"<!DOCTYPE html>",
		"xyztiles",
		"leaflet",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Response should contain %q", expected)
		}
	}
}

func TestHandleTileRequest_Success(t *testing.T) {
	srv := createTestServer(t)

	tests := []struct {
		path string
		name string
	}{
		{"/0/0/0.png", "zoom 0"},
		{"/1/0/0.png", "zoom 1 NW"},
		{"/1/1/1.png", "zoom 1 SE"},
		{"/2/1/1.png", "zoom 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.handleTileRequest(w, req, tt.path)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			// Check Content-Type
			contentType := resp.Header.Get("Content-Type")
			if contentType != "image/png" {
				t.Errorf("Expected Content-Type image/png, got %s", contentType)
			}

			// Check Cache-Control header
			cacheControl := resp.Header.Get("Cache-Control")
			if cacheControl != "public, max-age=86400" {
				t.Errorf("Expected Cache-Control 'public, max-age=86400', got %s", cacheControl)
			}

			// Verify it's a valid PNG
			img, err := png.Decode(resp.Body)
			if err != nil {
				t.Fatalf("Failed to decode PNG: %v", err)
			}

			// Check dimensions
			bounds := img.Bounds()
			if bounds.Dx() != 512 || bounds.Dy() != 512 {
				t.Errorf("Expected 512x512 image, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestHandleTileRequest_InvalidPath(t *testing.T) {
	srv := createTestServer(t)

	tests := []struct {
		path       string
		expectCode int
		name       string
	}{
		{"/invalid", http.StatusBadRequest, "invalid path"},
		{"/0/0/0.jpg", http.StatusBadRequest, "wrong extension"},
		{"/a/b/c.png", http.StatusBadRequest, "non-numeric coordinates"},
		{"/0/0.png", http.StatusBadRequest, "missing coordinate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.handleTileRequest(w, req, tt.path)

			resp := w.Result()
			if resp.StatusCode != tt.expectCode {
				t.Errorf("Expected status %d, got %d", tt.expectCode, resp.StatusCode)
			}
		})
	}
}

func TestHandleTileRequest_OutOfRange(t *testing.T) {
	srv := createTestServer(t)

	tests := []struct {
		path string
		name string
	}{
		{"/0/1/0.png", "x out of range for z0"},
		{"/0/0/1.png", "y out of range for z0"},
		{"/1/2/0.png", "x out of range for z1"},
		{"/1/0/2.png", "y out of range for z1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.handleTileRequest(w, req, tt.path)

			resp := w.Result()
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected status 404, got %d", resp.StatusCode)
			}
		})
	}
}

func TestHandler_Integration(t *testing.T) {
	srv := createTestServer(t)

	tests := []struct {
		path       string
		expectCode int
		expectType string
		name       string
	}{
		{"/", http.StatusOK, "text/html; charset=utf-8", "root"},
		{"/0/0/0.png", http.StatusOK, "image/png", "valid tile"},
		{"/1/0/0.png", http.StatusOK, "image/png", "valid tile z1"},
		{"/tile/0/0/0.png", http.StatusOK, "image/png", "/tile/ prefix"},
		{"/invalid/path", http.StatusBadRequest, "", "invalid path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.Handler().ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectCode {
				t.Errorf("Expected status %d, got %d", tt.expectCode, resp.StatusCode)
			}

			if tt.expectType != "" {
				contentType := resp.Header.Get("Content-Type")
				if contentType != tt.expectType {
					t.Errorf("Expected Content-Type %s, got %s", tt.expectType, contentType)
				}
			}
		})
	}
}

// createTestServer creates a server for testing
// Uses a small test image if the real image isn't available
func createTestServer(t *testing.T) *Server {
	t.Helper()

	var imagePath string

	// Check if test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		// Create a small test JPEG
		imagePath = createTestJPEG(t)
	} else {
		imagePath = testImagePath
	}

	cfg := Config{
		Port:      8080,
		ImagePath: imagePath,
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	return srv
}

// createTestJPEG creates a small test JPEG image for testing
func createTestJPEG(t *testing.T) string {
	t.Helper()

	// Create a small equirectangular test image (360x180)
	img := image.NewRGBA(image.Rect(0, 0, 360, 180))

	// Fill with a gradient
	for y := 0; y < 180; y++ {
		for x := 0; x < 360; x++ {
			c := color.RGBA{
				R: uint8(x * 255 / 360),
				G: uint8(y * 255 / 180),
				B: 128,
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	// Save to temp file
	tmpFile, err := os.CreateTemp("", "test-*.jpg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if err := jpeg.Encode(tmpFile, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}

	tmpFile.Close()

	// Schedule cleanup
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}

func TestTileEndpoint_RealWorld(t *testing.T) {
	// Check if test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image not found at %s, skipping test", testImagePath)
		return
	}

	srv := createTestServer(t)

	// Test a few tiles and save them to verify visually
	tiles := []struct {
		z, x, y int
	}{
		{0, 0, 0},
		{1, 0, 0},
		{2, 1, 1},
		{3, 4, 2},
	}

	tmpDir := os.TempDir()
	for _, tc := range tiles {
		path := filepath.Join("/", fmt.Sprint(tc.z), fmt.Sprint(tc.x), fmt.Sprint(tc.y)+".png")
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()

		srv.Handler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Tile %d/%d/%d failed with status %d", tc.z, tc.x, tc.y, resp.StatusCode)
			continue
		}

		// Optionally save the tile
		outputPath := filepath.Join(tmpDir, fmt.Sprintf("xyztiles_server_test_%d_%d_%d.png", tc.z, tc.x, tc.y))
		img, err := png.Decode(resp.Body)
		if err != nil {
			t.Errorf("Failed to decode tile %d/%d/%d: %v", tc.z, tc.x, tc.y, err)
			continue
		}

		outFile, err := os.Create(outputPath)
		if err != nil {
			t.Logf("Could not save test tile: %v", err)
			continue
		}

		png.Encode(outFile, img)
		outFile.Close()

		t.Logf("Saved tile %d/%d/%d to %s", tc.z, tc.x, tc.y, outputPath)
	}
}
