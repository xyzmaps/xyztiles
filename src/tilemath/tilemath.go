package tilemath

import (
	"fmt"
	"math"
)

// Bounds represents geographic bounds in decimal degrees (EPSG:4326)
type Bounds struct {
	West  float64 // Western longitude
	South float64 // Southern latitude
	East  float64 // Eastern longitude
	North float64 // Northern latitude
}

// TileCoord represents an XYZ tile coordinate
type TileCoord struct {
	Z int // Zoom level
	X int // Column (0 = left edge at -180°)
	Y int // Row (0 = top edge at ~85°N)
}

// MaxLatitude is the maximum latitude in Web Mercator projection (~85.0511°)
// This is the limit where the Mercator projection approaches infinity
const MaxLatitude = 85.05112878

// TileBounds calculates the geographic bounds of an XYZ tile.
// Returns bounds in EPSG:4326 (latitude/longitude in degrees).
func TileBounds(z, x, y int) (Bounds, error) {
	if z < 0 {
		return Bounds{}, fmt.Errorf("zoom level must be >= 0, got %d", z)
	}

	n := 1 << uint(z) // 2^z

	if x < 0 || x >= n {
		return Bounds{}, fmt.Errorf("x tile must be in range [0, %d) for zoom %d, got %d", n, z, x)
	}
	if y < 0 || y >= n {
		return Bounds{}, fmt.Errorf("y tile must be in range [0, %d) for zoom %d, got %d", n, z, y)
	}

	// Calculate bounds using tile corners
	// Top-left corner (x, y)
	west := tileXToLon(x, n)
	north := tileYToLat(y, n)

	// Bottom-right corner (x+1, y+1)
	east := tileXToLon(x+1, n)
	south := tileYToLat(y+1, n)

	return Bounds{
		West:  west,
		South: south,
		East:  east,
		North: north,
	}, nil
}

// tileXToLon converts a tile X coordinate to longitude in degrees
func tileXToLon(x, n int) float64 {
	return float64(x)/float64(n)*360.0 - 180.0
}

// tileYToLat converts a tile Y coordinate to latitude in degrees
// Uses the inverse Web Mercator projection formula
func tileYToLat(y, n int) float64 {
	// lat_rad = arctan(sinh(π * (1 - 2 * ytile / n)))
	ratio := 1.0 - 2.0*float64(y)/float64(n)
	latRad := math.Atan(math.Sinh(math.Pi * ratio))
	return latRad * 180.0 / math.Pi
}

// LonLatToTile converts longitude/latitude to the tile coordinate containing that point
func LonLatToTile(lon, lat float64, z int) (TileCoord, error) {
	if z < 0 {
		return TileCoord{}, fmt.Errorf("zoom level must be >= 0, got %d", z)
	}

	if lon < -180.0 || lon > 180.0 {
		return TileCoord{}, fmt.Errorf("longitude must be in range [-180, 180], got %f", lon)
	}

	// Clamp latitude to Web Mercator bounds
	if lat > MaxLatitude {
		lat = MaxLatitude
	}
	if lat < -MaxLatitude {
		lat = -MaxLatitude
	}

	n := 1 << uint(z) // 2^z
	nf := float64(n)

	// Convert longitude to x
	// Normalize longitude from [-180, 180] to [0, 1]
	x := (lon + 180.0) / 360.0
	xtile := int(math.Floor(nf * x))

	// Convert latitude to y using Web Mercator
	// y = arsinh(tan(lat)) = log[tan(lat) + sec(lat)]
	latRad := lat * math.Pi / 180.0
	y := math.Log(math.Tan(latRad) + 1.0/math.Cos(latRad))
	// Normalize to [0, 1] range
	y = (1.0 - y/math.Pi) / 2.0
	ytile := int(math.Floor(nf * y))

	// Clamp to valid tile range [0, n)
	if xtile >= n {
		xtile = n - 1
	}
	if ytile >= n {
		ytile = n - 1
	}
	if xtile < 0 {
		xtile = 0
	}
	if ytile < 0 {
		ytile = 0
	}

	return TileCoord{Z: z, X: xtile, Y: ytile}, nil
}

// String returns a string representation of the bounds
func (b Bounds) String() string {
	return fmt.Sprintf("Bounds[W:%.6f, S:%.6f, E:%.6f, N:%.6f]", b.West, b.South, b.East, b.North)
}

// String returns a string representation of the tile coordinate
func (tc TileCoord) String() string {
	return fmt.Sprintf("Tile[z:%d, x:%d, y:%d]", tc.Z, tc.X, tc.Y)
}
