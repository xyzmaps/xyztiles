package resources

import (
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
