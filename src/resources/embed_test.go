package resources

import (
	"strings"
	"testing"
)

func TestHasEmbeddedMap(t *testing.T) {
	hasMap := HasEmbeddedMap()
	if !hasMap {
		t.Error("Expected embedded map to be available")
	}
}

func TestDefaultMapSize(t *testing.T) {
	size := DefaultMapSize()
	if size <= 0 {
		t.Errorf("Expected positive map size, got %d", size)
	}

	// The test image should be around 1.6MB
	if size < 1000000 || size > 2000000 {
		t.Errorf("Expected map size around 1.6MB, got %d bytes", size)
	}

	t.Logf("Embedded map size: %d bytes", size)
}

func TestDefaultWorldMapNotNil(t *testing.T) {
	if DefaultWorldMap == nil {
		t.Error("DefaultWorldMap should not be nil")
	}

	if len(DefaultWorldMap) == 0 {
		t.Error("DefaultWorldMap should not be empty")
	}
}

func TestHasViewerHTML(t *testing.T) {
	hasViewer := HasViewerHTML()
	if !hasViewer {
		t.Error("Expected embedded viewer HTML to be available")
	}
}

func TestViewerHTML(t *testing.T) {
	if len(ViewerHTML) == 0 {
		t.Error("ViewerHTML should not be empty")
	}

	// Check for expected Leaflet-related content
	expectedStrings := []string{
		"<!DOCTYPE html>",
		"leaflet",
		"xyztiles",
		"<div id=\"map\">",
		"L.map",
		"L.tileLayer",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(ViewerHTML, expected) {
			t.Errorf("ViewerHTML should contain %q", expected)
		}
	}

	t.Logf("Viewer HTML size: %d bytes", len(ViewerHTML))
}
