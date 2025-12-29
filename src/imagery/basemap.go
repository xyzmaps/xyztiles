package imagery

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"os"

	"org.xyzmaps.xyztiles/src/tilemath"
	xdraw "golang.org/x/image/draw"
)

// BaseMap represents a loaded equirectangular world map image
// with methods to extract tiles. The image is kept in memory
// for fast tile generation.
type BaseMap struct {
	img    image.Image
	bounds image.Rectangle
	width  int
	height int
}

// TileSize is the output size for generated tiles (512x512 as per spec)
const TileSize = 512

// LoadJPEG loads a JPEG image from the given file path.
// The image is expected to be in equirectangular projection (EPSG:4326)
// covering the full world extent (-180, -90, 180, 90).
func LoadJPEG(path string) (*BaseMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JPEG: %w", err)
	}

	bounds := img.Bounds()
	return &BaseMap{
		img:    img,
		bounds: bounds,
		width:  bounds.Dx(),
		height: bounds.Dy(),
	}, nil
}

// LoadJPEGFromBytes loads a JPEG image from a byte slice (e.g., embedded resource).
// The image is expected to be in equirectangular projection (EPSG:4326)
// covering the full world extent (-180, -90, 180, 90).
func LoadJPEGFromBytes(data []byte) (*BaseMap, error) {
	reader := bytes.NewReader(data)

	img, err := jpeg.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JPEG from bytes: %w", err)
	}

	bounds := img.Bounds()
	return &BaseMap{
		img:    img,
		bounds: bounds,
		width:  bounds.Dx(),
		height: bounds.Dy(),
	}, nil
}

// ExtractTile extracts and resamples a tile region from the base map.
// Returns a 512x512 RGBA image containing the tile at the given XYZ coordinates.
func (bm *BaseMap) ExtractTile(z, x, y int) (*image.RGBA, error) {
	// Get geographic bounds of the tile
	tileBounds, err := tilemath.TileBounds(z, x, y)
	if err != nil {
		return nil, fmt.Errorf("invalid tile coordinates: %w", err)
	}

	// Convert geographic bounds to pixel bounds in the source image
	pixelBounds := bm.geoBoundsToPixelBounds(tileBounds)

	// Extract the source region
	sourceRegion := bm.extractRegion(pixelBounds)

	// Resample to 512x512 using CatmullRom interpolation for better quality
	tile := image.NewRGBA(image.Rect(0, 0, TileSize, TileSize))
	xdraw.CatmullRom.Scale(tile, tile.Bounds(), sourceRegion, sourceRegion.Bounds(), xdraw.Over, nil)

	return tile, nil
}

// geoBoundsToPixelBounds converts geographic bounds (lat/lon) to pixel bounds
// in the equirectangular source image.
// For equirectangular projection covering full world extent:
//   pixel_x = (lon + 180) / 360 * image_width
//   pixel_y = (90 - lat) / 180 * image_height
func (bm *BaseMap) geoBoundsToPixelBounds(geo tilemath.Bounds) image.Rectangle {
	// Convert west/east longitude to x coordinates
	x0 := lonToPixelX(geo.West, bm.width)
	x1 := lonToPixelX(geo.East, bm.width)

	// Convert north/south latitude to y coordinates
	// Note: north latitude maps to smaller y (top of image)
	y0 := latToPixelY(geo.North, bm.height)
	y1 := latToPixelY(geo.South, bm.height)

	// Clamp to image bounds
	x0 = clamp(x0, 0, bm.width)
	x1 = clamp(x1, 0, bm.width)
	y0 = clamp(y0, 0, bm.height)
	y1 = clamp(y1, 0, bm.height)

	return image.Rect(x0, y0, x1, y1)
}

// extractRegion extracts a sub-image from the base map.
// For efficiency, this uses SubImage if available, otherwise copies the region.
func (bm *BaseMap) extractRegion(bounds image.Rectangle) image.Image {
	// Check if we can use SubImage (most image types support this)
	if subber, ok := bm.img.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		return subber.SubImage(bounds)
	}

	// Fallback: copy the region
	region := image.NewRGBA(bounds)
	draw.Draw(region, bounds, bm.img, bounds.Min, draw.Src)
	return region
}

// lonToPixelX converts longitude to pixel x coordinate
func lonToPixelX(lon float64, imageWidth int) int {
	// Normalize longitude from [-180, 180] to [0, 1]
	normalized := (lon + 180.0) / 360.0
	return int(normalized * float64(imageWidth))
}

// latToPixelY converts latitude to pixel y coordinate
func latToPixelY(lat float64, imageHeight int) int {
	// Normalize latitude from [90, -90] to [0, 1]
	// Note: y increases downward in images
	normalized := (90.0 - lat) / 180.0
	return int(normalized * float64(imageHeight))
}

// clamp restricts a value to the range [min, max]
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Width returns the width of the base map image
func (bm *BaseMap) Width() int {
	return bm.width
}

// Height returns the height of the base map image
func (bm *BaseMap) Height() int {
	return bm.height
}
