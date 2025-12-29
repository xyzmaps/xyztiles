package tilemath

import (
	"math"
	"testing"
)

func TestTileBounds_Zoom0(t *testing.T) {
	// Zoom 0 has only one tile covering the entire world
	bounds, err := TileBounds(0, 0, 0)
	if err != nil {
		t.Fatalf("TileBounds(0, 0, 0) failed: %v", err)
	}

	// Should span full longitude and Web Mercator latitude range
	assertFloat64Near(t, -180.0, bounds.West, 1e-6, "zoom 0 west")
	assertFloat64Near(t, 180.0, bounds.East, 1e-6, "zoom 0 east")
	assertFloat64Near(t, -MaxLatitude, bounds.South, 1e-6, "zoom 0 south")
	assertFloat64Near(t, MaxLatitude, bounds.North, 1e-6, "zoom 0 north")
}

func TestTileBounds_Zoom1(t *testing.T) {
	// Zoom 1 has 2x2 = 4 tiles
	tests := []struct {
		x, y        int
		westExpect  float64
		southExpect float64
		eastExpect  float64
		northExpect float64
		name        string
	}{
		{0, 0, -180.0, 0.0, 0.0, MaxLatitude, "top-left (NW quadrant)"},
		{1, 0, 0.0, 0.0, 180.0, MaxLatitude, "top-right (NE quadrant)"},
		{0, 1, -180.0, -MaxLatitude, 0.0, 0.0, "bottom-left (SW quadrant)"},
		{1, 1, 0.0, -MaxLatitude, 180.0, 0.0, "bottom-right (SE quadrant)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bounds, err := TileBounds(1, tt.x, tt.y)
			if err != nil {
				t.Fatalf("TileBounds(1, %d, %d) failed: %v", tt.x, tt.y, err)
			}

			assertFloat64Near(t, tt.westExpect, bounds.West, 1e-6, "west")
			assertFloat64Near(t, tt.eastExpect, bounds.East, 1e-6, "east")
			assertFloat64Near(t, tt.southExpect, bounds.South, 1e-6, "south")
			assertFloat64Near(t, tt.northExpect, bounds.North, 1e-6, "north")
		})
	}
}

func TestTileBounds_SpecificTiles(t *testing.T) {
	tests := []struct {
		z, x, y     int
		westExpect  float64
		southExpect float64
		eastExpect  float64
		northExpect float64
		name        string
	}{
		// Zoom 2, center tile should straddle equator and prime meridian
		{2, 2, 2, 0.0, -66.51326, 90.0, 0.0, "z2 center-right"},

		// Known tile from OpenStreetMap
		// Zoom 4, tile containing London (approximately)
		{4, 7, 5, -22.5, 40.97989, 0.0, 55.77657, "z4 tile near London"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bounds, err := TileBounds(tt.z, tt.x, tt.y)
			if err != nil {
				t.Fatalf("TileBounds(%d, %d, %d) failed: %v", tt.z, tt.x, tt.y, err)
			}

			assertFloat64Near(t, tt.westExpect, bounds.West, 1e-4, "west")
			assertFloat64Near(t, tt.eastExpect, bounds.East, 1e-4, "east")
			assertFloat64Near(t, tt.southExpect, bounds.South, 1e-4, "south")
			assertFloat64Near(t, tt.northExpect, bounds.North, 1e-4, "north")
		})
	}
}

func TestTileBounds_Errors(t *testing.T) {
	tests := []struct {
		z, x, y int
		name    string
	}{
		{-1, 0, 0, "negative zoom"},
		{0, 1, 0, "x out of range for z0"},
		{0, 0, 1, "y out of range for z0"},
		{1, -1, 0, "negative x"},
		{1, 0, -1, "negative y"},
		{2, 4, 0, "x too large for z2"},
		{2, 0, 4, "y too large for z2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := TileBounds(tt.z, tt.x, tt.y)
			if err == nil {
				t.Errorf("TileBounds(%d, %d, %d) should return error but got nil", tt.z, tt.x, tt.y)
			}
		})
	}
}

func TestLonLatToTile_Zoom0(t *testing.T) {
	// Any point should map to tile (0, 0) at zoom 0
	tests := []struct {
		lon, lat float64
		name     string
	}{
		{0, 0, "origin"},
		{-180, 0, "west edge"},
		{180, 0, "east edge"},
		{0, MaxLatitude, "north edge"},
		{0, -MaxLatitude, "south edge"},
		{-122.4, 37.8, "San Francisco"},
		{2.35, 48.86, "Paris"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile, err := LonLatToTile(tt.lon, tt.lat, 0)
			if err != nil {
				t.Fatalf("LonLatToTile(%f, %f, 0) failed: %v", tt.lon, tt.lat, err)
			}

			if tile.Z != 0 || tile.X != 0 || tile.Y != 0 {
				t.Errorf("Expected tile (0,0,0), got (%d,%d,%d)", tile.Z, tile.X, tile.Y)
			}
		})
	}
}

func TestLonLatToTile_Zoom1(t *testing.T) {
	tests := []struct {
		lon, lat   float64
		xExpect    int
		yExpect    int
		name       string
	}{
		{-90, 45, 0, 0, "northwest quadrant"},
		{90, 45, 1, 0, "northeast quadrant"},
		{-90, -45, 0, 1, "southwest quadrant"},
		{90, -45, 1, 1, "southeast quadrant"},
		{0, 0, 1, 1, "origin (equator/prime meridian)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile, err := LonLatToTile(tt.lon, tt.lat, 1)
			if err != nil {
				t.Fatalf("LonLatToTile(%f, %f, 1) failed: %v", tt.lon, tt.lat, err)
			}

			if tile.X != tt.xExpect || tile.Y != tt.yExpect {
				t.Errorf("Expected tile (1,%d,%d), got (1,%d,%d)",
					tt.xExpect, tt.yExpect, tile.X, tile.Y)
			}
		})
	}
}

func TestLonLatToTile_LatitudeClamping(t *testing.T) {
	// Latitudes beyond Web Mercator bounds should be clamped
	tests := []struct {
		lat  float64
		name string
	}{
		{90.0, "north pole"},
		{-90.0, "south pole"},
		{100.0, "beyond north"},
		{-100.0, "beyond south"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LonLatToTile(0, tt.lat, 5)
			if err != nil {
				t.Errorf("LonLatToTile(0, %f, 5) should not error with lat clamping, got: %v", tt.lat, err)
			}
		})
	}
}

func TestLonLatToTile_Errors(t *testing.T) {
	tests := []struct {
		lon, lat float64
		z        int
		name     string
	}{
		{0, 0, -1, "negative zoom"},
		{-181, 0, 5, "longitude too small"},
		{181, 0, 5, "longitude too large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LonLatToTile(tt.lon, tt.lat, tt.z)
			if err == nil {
				t.Errorf("LonLatToTile(%f, %f, %d) should return error but got nil", tt.lon, tt.lat, tt.z)
			}
		})
	}
}

func TestRoundTrip_LonLatToTileAndBack(t *testing.T) {
	// Test that converting lon/lat -> tile -> bounds includes the original point
	tests := []struct {
		lon, lat float64
		z        int
		name     string
	}{
		{0, 0, 5, "origin at z5"},
		{-122.4, 37.8, 10, "San Francisco at z10"},
		{2.35, 48.86, 10, "Paris at z10"},
		{139.69, 35.68, 10, "Tokyo at z10"},
		{-43.2, -22.9, 10, "Rio at z10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert point to tile
			tile, err := LonLatToTile(tt.lon, tt.lat, tt.z)
			if err != nil {
				t.Fatalf("LonLatToTile failed: %v", err)
			}

			// Get bounds of that tile
			bounds, err := TileBounds(tile.Z, tile.X, tile.Y)
			if err != nil {
				t.Fatalf("TileBounds failed: %v", err)
			}

			// Original point should be within tile bounds
			if tt.lon < bounds.West || tt.lon > bounds.East {
				t.Errorf("Longitude %f not in bounds [%f, %f]", tt.lon, bounds.West, bounds.East)
			}
			if tt.lat < bounds.South || tt.lat > bounds.North {
				t.Errorf("Latitude %f not in bounds [%f, %f]", tt.lat, bounds.South, bounds.North)
			}
		})
	}
}

func TestTileXToLon(t *testing.T) {
	tests := []struct {
		x, n   int
		expect float64
		name   string
	}{
		{0, 1, -180.0, "z0 left edge"},
		{1, 1, 180.0, "z0 right edge"},
		{0, 2, -180.0, "z1 left edge"},
		{1, 2, 0.0, "z1 center"},
		{2, 2, 180.0, "z1 right edge"},
		{2, 4, 0.0, "z2 center"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lon := tileXToLon(tt.x, tt.n)
			assertFloat64Near(t, tt.expect, lon, 1e-6, "longitude")
		})
	}
}

func TestTileYToLat(t *testing.T) {
	tests := []struct {
		y, n   int
		expect float64
		name   string
	}{
		{0, 1, MaxLatitude, "z0 top edge"},
		{1, 1, -MaxLatitude, "z0 bottom edge"},
		{0, 2, MaxLatitude, "z1 top edge"},
		{1, 2, 0.0, "z1 center"},
		{2, 2, -MaxLatitude, "z1 bottom edge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat := tileYToLat(tt.y, tt.n)
			assertFloat64Near(t, tt.expect, lat, 1e-6, "latitude")
		})
	}
}

func TestBounds_String(t *testing.T) {
	bounds := Bounds{West: -122.5, South: 37.5, East: -122.0, North: 38.0}
	str := bounds.String()
	expected := "Bounds[W:-122.500000, S:37.500000, E:-122.000000, N:38.000000]"
	if str != expected {
		t.Errorf("Expected %q, got %q", expected, str)
	}
}

func TestTileCoord_String(t *testing.T) {
	tile := TileCoord{Z: 5, X: 10, Y: 15}
	str := tile.String()
	expected := "Tile[z:5, x:10, y:15]"
	if str != expected {
		t.Errorf("Expected %q, got %q", expected, str)
	}
}

// assertFloat64Near checks if two float64 values are within epsilon of each other
func assertFloat64Near(t *testing.T, expected, actual, epsilon float64, name string) {
	t.Helper()
	if math.Abs(expected-actual) > epsilon {
		t.Errorf("%s: expected %.10f, got %.10f (diff: %.10e)",
			name, expected, actual, math.Abs(expected-actual))
	}
}
