# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xyztiles is a Go-based web map tile server that serves XYZ tiles from an embedded or custom equirectangular world map image. The application is a single binary with zero runtime dependencies, perfect for educational purposes and offline environments.

**Module**: `org.xyzmaps.xyztiles`
**Go Version**: 1.25.5
**Binary Size**: ~13MB (includes embedded 1.6MB JPEG + Leaflet viewer)

## Key Features

- Embedded NASA Blue Marble world map (5400×2700 JPEG)
- Interactive Leaflet viewer with debug mode
- On-demand tile generation with CatmullRom interpolation
- Standard XYZ tile endpoint (`/{z}/{x}/{y}.png`)
- Web Mercator projection (EPSG:3857)
- High test coverage (97%+ across core modules)

## Project Structure

```
xyztiles/
├── cmd/                    # CLI commands using Cobra
│   ├── root.go            # Main command with server logic
│   └── update.go          # Auto-update command
├── src/
│   ├── imagery/           # Image loading and tile generation
│   │   ├── basemap.go     # BaseMap struct, JPEG loading, tile extraction
│   │   └── basemap_test.go
│   ├── resources/         # Embedded assets
│   │   ├── embed.go       # Go embed directives for JPEG + HTML
│   │   ├── viewer.html    # Leaflet viewer (embedded at compile-time)
│   │   └── world.topo.200407.3x5400x2700.jpg (embedded)
│   ├── server/            # HTTP server and handlers
│   │   ├── server.go      # Server struct, tile endpoint, viewer endpoint
│   │   └── server_test.go
│   ├── tilemath/          # Coordinate transformations
│   │   ├── tilemath.go    # XYZ ↔ lat/lon conversions (Web Mercator)
│   │   └── tilemath_test.go
│   └── version/           # Version information
├── res/                   # Source resources (not embedded, for development)
│   └── world.topo.200407.3x5400x2700.jpg
├── main.go                # Entry point
├── SPEC.md                # Project specification
└── README.md              # User documentation
```

## Build and Development Commands

### Local Development

```bash
# Build the application
go build -o xyztiles main.go

# Run with embedded map
./xyztiles

# Run with custom image
./xyztiles --image path/to/custom.jpg --port 8080

# Format code
go fmt ./...

# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with verbose output
go test -v ./src/...

# Run specific package tests
go test ./src/tilemath -v
go test ./src/imagery -v
go test ./src/server -v

# Tidy dependencies
go mod tidy
```

### Testing Individual Components

```bash
# Test tile math (coordinate conversions)
go test ./src/tilemath -run TestTileBounds

# Test image loading and extraction
go test ./src/imagery -run TestExtractTile

# Test HTTP server
go test ./src/server -run TestHandleTileRequest
```

### Release Building

```bash
# Build release artifacts for all platforms (Linux, Windows, macOS)
goreleaser build --snapshot --clean

# Create a full release (requires tags and proper git state)
goreleaser release --snapshot --clean

# Test the release configuration without building
goreleaser check
```

## Architecture

### Data Flow

```
HTTP Request /{z}/{x}/{y}.png
    ↓
Server.handleTileRequest()
    ↓
tilemath.TileBounds(z, x, y) → Geographic bounds
    ↓
imagery.BaseMap.ExtractTile(z, x, y)
    ↓
    1. Convert geo bounds to pixel bounds (equirectangular)
    2. Extract region from source image
    3. Resample to 512×512 (CatmullRom interpolation)
    ↓
PNG encoding + HTTP response with cache headers
```

### Key Components

#### `src/tilemath`
- **Pure math functions** for XYZ tile coordinate conversions
- **Web Mercator projection** formulas (±85.0511° latitude limit)
- **Functions**: `TileBounds()`, `LonLatToTile()`
- **No dependencies** on other packages

#### `src/imagery`
- **Image loading** from file or embedded bytes
- **Tile extraction** with coordinate transformation
- **Resampling** using `golang.org/x/image/draw` (CatmullRom)
- **Supports**: JPEG input (PNG/TIFF planned)

#### `src/server`
- **HTTP server** using standard `net/http`
- **Tile endpoint**: `/{z}/{x}/{y}.png` with proper cache headers
- **Viewer endpoint**: `/` serving embedded Leaflet HTML
- **Error handling**: 404 for invalid tiles, 400 for bad requests

#### `src/resources`
- **Embedded assets** using `//go:embed`
- **DefaultWorldMap**: 1.6MB JPEG (5400×2700 NASA Blue Marble)
- **ViewerHTML**: Interactive Leaflet map with debug mode
- **Compile-time embedding** - no external files needed at runtime

## Important Implementation Details

### Coordinate Systems

1. **Input Image**: Equirectangular (EPSG:4326)
   - Linear mapping: `pixel_x = (lon + 180) / 360 * width`
   - Linear mapping: `pixel_y = (90 - lat) / 180 * height`

2. **Output Tiles**: Web Mercator (EPSG:3857)
   - Non-linear latitude projection
   - Formulas in `src/tilemath/tilemath.go`

### Image Resampling

- **Algorithm**: CatmullRom interpolation (changed from BiLinear for better quality)
- **Tile Size**: 512×512 pixels (configured in `imagery.TileSize`)
- **Zoom Levels**: 0-6 native, 7-10 browser-scaled
- **Location**: `src/imagery/basemap.go:89`

### Embedded Resources

Resources are embedded at **compile-time** using `//go:embed` directives:

```go
//go:embed world.topo.200407.3x5400x2700.jpg
var DefaultWorldMap []byte

//go:embed viewer.html
var ViewerHTML string
```

This means:
- Changes to `src/resources/*.{jpg,html}` require **rebuilding** the binary
- The `res/` directory is for **development only** (source files)
- The `src/resources/` directory contains the **embedded copies**

## Development Workflow

### Adding Features

1. Write tests first (TDD approach preferred)
2. Implement feature in appropriate package
3. Update tests to maintain >95% coverage
4. Run full test suite: `go test ./...`
5. Update documentation if needed

### Modifying Embedded Resources

1. Edit the source file in `src/resources/`
2. **Rebuild the binary**: `go build -o xyztiles main.go`
3. Test the changes: `./xyztiles`
4. The new resources are now embedded

### Testing Strategy

- **Unit tests**: All packages have `*_test.go` files
- **Integration tests**: Server tests include full HTTP request/response testing
- **Coverage target**: >95% for core logic packages
- **Test data**: Uses real JPEG in `res/` directory

## Common Tasks

### Update the Embedded World Map

```bash
# 1. Replace the source image
cp new-world-map.jpg src/resources/world.topo.200407.3x5400x2700.jpg

# 2. Rebuild (embeds the new image)
go build -o xyztiles main.go

# 3. Test
./xyztiles
```

### Modify the Leaflet Viewer

```bash
# 1. Edit the HTML
vim src/resources/viewer.html

# 2. Rebuild (embeds the new HTML)
go build -o xyztiles main.go

# 3. Test
./xyztiles
# Open http://localhost:8080 in browser
```

### Add New Tile Math Functions

```bash
# 1. Add function to src/tilemath/tilemath.go
# 2. Add tests to src/tilemath/tilemath_test.go
# 3. Run tests
go test ./src/tilemath -v

# 4. Check coverage
go test ./src/tilemath -cover
```

## Dependencies

### Runtime (Embedded in Binary)
- **Zero external dependencies** - completely self-contained

### Build-time Only
- `golang.org/x/image/draw` - Image resampling algorithms
- `github.com/spf13/cobra` - CLI framework

### External (Loaded by Browser)
- Leaflet CSS/JS from unpkg.com CDN (in viewer HTML)
- Only needed when viewing the embedded map interface

## Performance Considerations

- **Tile Generation**: ~10-200ms depending on zoom and image complexity
- **Memory**: ~50MB base + size of loaded image
- **Caching**: Currently no caching (tiles generated on each request)
  - **Future**: LRU cache planned to speed up repeat requests
- **Concurrency**: Safe for concurrent tile requests

## Known Limitations

1. **Image Format**: Only JPEG input currently supported
2. **Projection**: Only equirectangular input images
3. **Caching**: No in-memory cache yet (generates tiles on every request)
4. **Zoom Levels**: Only pre-generates tiles up to zoom 6
5. **Interpolation**: CatmullRom is high quality but slower than bilinear

## Troubleshooting

### Binary Size Too Large
- The 13MB size includes the 1.6MB embedded JPEG
- Use a smaller source image if needed
- The size is reasonable for a self-contained application

### Tests Failing
```bash
# Check if the test image exists
ls -lh res/world.topo.200407.3x5400x2700.jpg
ls -lh src/resources/world.topo.200407.3x5400x2700.jpg

# Run tests with verbose output to see details
go test -v ./...
```

### Viewer Not Loading
- Check browser console for errors
- Verify server logs show tile requests
- Try the debug mode (press 'D' key in viewer)

## Resources

- **Spec**: See `SPEC.md` for original requirements
- **OSM Tile Math**: https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
- **Leaflet Docs**: https://leafletjs.com/reference.html
- **NASA Blue Marble**: https://visibleearth.nasa.gov/collection/1484/blue-marble
