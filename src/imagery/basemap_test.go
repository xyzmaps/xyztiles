package imagery

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"org.xyzmaps.xyztiles/src/tilemath"
)

const testImagePath = "../../res/world.topo.200407.3x5400x2700.jpg"

func TestLoadJPEG(t *testing.T) {
	// Check if test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image not found at %s, skipping test", testImagePath)
		return
	}

	basemap, err := LoadJPEG(testImagePath)
	if err != nil {
		t.Fatalf("LoadJPEG failed: %v", err)
	}

	if basemap == nil {
		t.Fatal("LoadJPEG returned nil basemap")
	}

	// Verify dimensions are reasonable (should be 5400x2700 based on filename)
	if basemap.Width() <= 0 || basemap.Height() <= 0 {
		t.Errorf("Invalid dimensions: %dx%d", basemap.Width(), basemap.Height())
	}

	t.Logf("Loaded basemap: %dx%d", basemap.Width(), basemap.Height())
}

func TestLoadJPEG_InvalidPath(t *testing.T) {
	_, err := LoadJPEG("/nonexistent/path/image.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadJPEG_InvalidImage(t *testing.T) {
	// Create a temporary non-JPEG file
	tmpFile, err := os.CreateTemp("", "invalid-*.jpg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some non-JPEG data
	tmpFile.WriteString("This is not a JPEG image")
	tmpFile.Close()

	_, err = LoadJPEG(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid JPEG, got nil")
	}
}

func TestLonToPixelX(t *testing.T) {
	tests := []struct {
		lon        float64
		imageWidth int
		expected   int
		name       string
	}{
		{-180.0, 3600, 0, "west edge"},
		{180.0, 3600, 3600, "east edge"},
		{0.0, 3600, 1800, "prime meridian"},
		{-90.0, 3600, 900, "90W"},
		{90.0, 3600, 2700, "90E"},
		{-180.0, 5400, 0, "west edge 5400px"},
		{0.0, 5400, 2700, "prime meridian 5400px"},
		{180.0, 5400, 5400, "east edge 5400px"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lonToPixelX(tt.lon, tt.imageWidth)
			if result != tt.expected {
				t.Errorf("lonToPixelX(%f, %d) = %d, expected %d",
					tt.lon, tt.imageWidth, result, tt.expected)
			}
		})
	}
}

func TestLatToPixelY(t *testing.T) {
	tests := []struct {
		lat         float64
		imageHeight int
		expected    int
		name        string
	}{
		{90.0, 1800, 0, "north pole"},
		{-90.0, 1800, 1800, "south pole"},
		{0.0, 1800, 900, "equator"},
		{45.0, 1800, 450, "45N"},
		{-45.0, 1800, 1350, "45S"},
		{90.0, 2700, 0, "north pole 2700px"},
		{0.0, 2700, 1350, "equator 2700px"},
		{-90.0, 2700, 2700, "south pole 2700px"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := latToPixelY(tt.lat, tt.imageHeight)
			if result != tt.expected {
				t.Errorf("latToPixelY(%f, %d) = %d, expected %d",
					tt.lat, tt.imageHeight, result, tt.expected)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected int
		name                      string
	}{
		{5, 0, 10, 5, "within range"},
		{-5, 0, 10, 0, "below min"},
		{15, 0, 10, 10, "above max"},
		{0, 0, 10, 0, "at min"},
		{10, 0, 10, 10, "at max"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clamp(%d, %d, %d) = %d, expected %d",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestGeoBoundsToPixelBounds(t *testing.T) {
	// Create a test basemap with known dimensions
	basemap := &BaseMap{
		img:    nil, // not needed for this test
		bounds: image.Rect(0, 0, 3600, 1800),
		width:  3600,
		height: 1800,
	}

	tests := []struct {
		geo      tilemath.Bounds
		expected image.Rectangle
		name     string
	}{
		{
			tilemath.Bounds{West: -180, South: -90, East: 180, North: 90},
			image.Rect(0, 0, 3600, 1800),
			"full world",
		},
		{
			tilemath.Bounds{West: -180, South: 0, East: 0, North: 90},
			image.Rect(0, 0, 1800, 900),
			"northwest quadrant",
		},
		{
			tilemath.Bounds{West: 0, South: 0, East: 180, North: 90},
			image.Rect(1800, 0, 3600, 900),
			"northeast quadrant",
		},
		{
			tilemath.Bounds{West: 0, South: -90, East: 180, North: 0},
			image.Rect(1800, 900, 3600, 1800),
			"southeast quadrant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := basemap.geoBoundsToPixelBounds(tt.geo)
			if result != tt.expected {
				t.Errorf("geoBoundsToPixelBounds(%v) = %v, expected %v",
					tt.geo, result, tt.expected)
			}
		})
	}
}

func TestExtractTile(t *testing.T) {
	// Check if test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image not found at %s, skipping test", testImagePath)
		return
	}

	basemap, err := LoadJPEG(testImagePath)
	if err != nil {
		t.Fatalf("LoadJPEG failed: %v", err)
	}

	tests := []struct {
		z, x, y int
		name    string
	}{
		{0, 0, 0, "zoom 0 (entire world)"},
		{1, 0, 0, "zoom 1 northwest"},
		{1, 1, 0, "zoom 1 northeast"},
		{1, 0, 1, "zoom 1 southwest"},
		{1, 1, 1, "zoom 1 southeast"},
		{2, 1, 1, "zoom 2 sample tile"},
		{3, 4, 2, "zoom 3 sample tile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tile, err := basemap.ExtractTile(tt.z, tt.x, tt.y)
			if err != nil {
				t.Fatalf("ExtractTile(%d, %d, %d) failed: %v", tt.z, tt.x, tt.y, err)
			}

			if tile == nil {
				t.Fatal("ExtractTile returned nil tile")
			}

			// Verify tile is exactly 512x512
			bounds := tile.Bounds()
			if bounds.Dx() != TileSize || bounds.Dy() != TileSize {
				t.Errorf("Expected %dx%d tile, got %dx%d",
					TileSize, TileSize, bounds.Dx(), bounds.Dy())
			}

			// Verify tile is not completely transparent or empty
			hasNonZeroPixel := false
			for y := bounds.Min.Y; y < bounds.Max.Y && !hasNonZeroPixel; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, a := tile.At(x, y).RGBA()
					if r > 0 || g > 0 || b > 0 || a > 0 {
						hasNonZeroPixel = true
						break
					}
				}
			}
			if !hasNonZeroPixel {
				t.Error("Tile appears to be completely empty")
			}

			t.Logf("Successfully extracted tile %d/%d/%d (%dx%d)",
				tt.z, tt.x, tt.y, bounds.Dx(), bounds.Dy())
		})
	}
}

func TestExtractTile_InvalidCoordinates(t *testing.T) {
	// Create a minimal test image
	testImg := createTestImage(100, 50)
	basemap := &BaseMap{
		img:    testImg,
		bounds: testImg.Bounds(),
		width:  100,
		height: 50,
	}

	tests := []struct {
		z, x, y int
		name    string
	}{
		{-1, 0, 0, "negative zoom"},
		{0, 1, 0, "x out of range for z0"},
		{0, 0, 1, "y out of range for z0"},
		{1, -1, 0, "negative x"},
		{1, 0, -1, "negative y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := basemap.ExtractTile(tt.z, tt.x, tt.y)
			if err == nil {
				t.Errorf("ExtractTile(%d, %d, %d) should return error but got nil",
					tt.z, tt.x, tt.y)
			}
		})
	}
}

func TestExtractTile_SaveSampleTile(t *testing.T) {
	// Check if test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image not found at %s, skipping test", testImagePath)
		return
	}

	basemap, err := LoadJPEG(testImagePath)
	if err != nil {
		t.Fatalf("LoadJPEG failed: %v", err)
	}

	// Extract a sample tile (zoom 2, roughly centered)
	tile, err := basemap.ExtractTile(2, 2, 1)
	if err != nil {
		t.Fatalf("ExtractTile failed: %v", err)
	}

	// Save to a test output file (in a temp directory)
	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, "xyztiles_test_tile_2_2_1.jpg")

	outFile, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, tile, &jpeg.Options{Quality: 85})
	if err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}

	t.Logf("Sample tile saved to: %s", outputPath)
}

// createTestImage creates a simple test image with a gradient
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a simple gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 128,
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	return img
}

func TestExtractRegion(t *testing.T) {
	// Create a test image
	testImg := createTestImage(100, 50)
	basemap := &BaseMap{
		img:    testImg,
		bounds: testImg.Bounds(),
		width:  100,
		height: 50,
	}

	// Extract a region
	region := basemap.extractRegion(image.Rect(10, 10, 30, 30))
	if region == nil {
		t.Fatal("extractRegion returned nil")
	}

	// Verify region has correct bounds
	bounds := region.Bounds()
	if bounds.Dx() != 20 || bounds.Dy() != 20 {
		t.Errorf("Expected 20x20 region, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestExtractRegion_WithoutSubImageSupport(t *testing.T) {
	// Create a test image that doesn't support SubImage
	testImg := createTestImage(100, 50)

	// Wrap it in a type that doesn't implement SubImage
	wrappedImg := &imageWrapper{Image: testImg}

	basemap := &BaseMap{
		img:    wrappedImg,
		bounds: testImg.Bounds(),
		width:  100,
		height: 50,
	}

	// Extract a region (should use fallback copy method)
	region := basemap.extractRegion(image.Rect(10, 10, 30, 30))
	if region == nil {
		t.Fatal("extractRegion returned nil")
	}

	// Verify region exists and has correct bounds
	bounds := region.Bounds()
	if bounds.Dx() != 20 || bounds.Dy() != 20 {
		t.Errorf("Expected 20x20 region, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// imageWrapper wraps an image.Image without implementing SubImage
type imageWrapper struct {
	image.Image
}

func (w *imageWrapper) ColorModel() color.Model {
	return w.Image.ColorModel()
}

func (w *imageWrapper) Bounds() image.Rectangle {
	return w.Image.Bounds()
}

func (w *imageWrapper) At(x, y int) color.Color {
	return w.Image.At(x, y)
}
