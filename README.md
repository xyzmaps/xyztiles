# xyztiles

> A single Go binary that serves web map tiles from an embedded or custom equirectangular world map image.

**Zero external dependencies, no hosted services required** - perfect for learning web mapping, local development, or offline/air-gapped environments.

![Go Version](https://img.shields.io/badge/go-1.25.5-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## Features

- 🌍 **Embedded World Map** - NASA Blue Marble imagery (5400×2700) included in the binary
- 🗺️ **Interactive Leaflet Viewer** - Prototype web map interface with debug mode
- 🚀 **Zero Configuration** - Just run the binary and open your browser
- 📦 **Single Binary** - 13MB standalone executable (includes map + viewer)
- 🎨 **Custom Images** - Support for your own equirectangular JPEG images
- ⚡ **On-Demand Tile Generation** - Tiles created in real-time with CatmullRom interpolation
- 🔧 **Standard XYZ Tiles** - Compatible with XYZmaps and tools like OpenStreetMap, Leaflet, and other web mapping libraries

## Quick Start

```bash
# Download and run (or build from source)
./xyztiles

# Server starts on http://localhost:8080
# Open your browser to see the interactive map viewer!
```

That's it! The embedded map and viewer are ready to go.

## Installation

### Option 1: Download Release Binary

```bash
# Download the latest release for your platform
# (Replace with actual release URL when available)
curl -L -o xyztiles https://github.com/xyzmaps/xyztiles/releases/latest/download/xyztiles-linux-amd64
chmod +x xyztiles
./xyztiles
```

### Option 2: Build from Source

**Requirements:**
- Go 1.25.5 or later
- No other dependencies!

```bash
# Clone the repository
git clone https://github.com/xyzmaps/xyztiles.git
cd xyztiles

# Build the binary
go build -o xyztiles main.go

# Run it
./xyztiles
```

## Usage

### Basic Usage (Embedded Map)

```bash
# Start server with embedded world map on default port 8080
./xyztiles

# Custom port
./xyztiles --port 9000
```

Then open your browser to `http://localhost:8080` (or your custom port) to see the interactive map viewer.

### Using a Custom Image

```bash
# Use your own equirectangular world map image
./xyztiles --image path/to/your/worldmap.jpg --port 8080
```

**Image Requirements:**
- Format: JPEG (PNG and TIFF support coming soon)
- Projection: Equirectangular (EPSG:4326)
- Coverage: Full world extent (-180°, -90°, 180°, 90°)
- Example: NASA Blue Marble, Natural Earth, custom satellite imagery

### CLI Options

```
Flags:
  -h, --help           help for xyztiles
  -i, --image string   Path to custom equirectangular world map image
                       (optional, uses embedded map if not specified)
  -p, --port int       Port to run the server on (default 8080)
  -v, --version        Print version information
```

## Tile Endpoint

Tiles are served at the standard XYZ URL pattern:

```
/{z}/{x}/{y}.png
```

**Examples:**
- `http://localhost:8080/0/0/0.png` - Zoom 0 (entire world)
- `http://localhost:8080/1/0/0.png` - Zoom 1, northwest quadrant
- `http://localhost:8080/6/32/21.png` - Zoom 6, specific tile

**Tile Specifications:**
- Size: 512×512 pixels (256 pixel at 2x/retina)
- Format: PNG
- Projection: Web Mercator (EPSG:3857)
- Zoom Levels: 0-6 (native), 7-10 (browser-scaled)
- Interpolation: CatmullRom for high quality
- Cache Headers: 24 hours (`max-age=86400`)

## Using with Leaflet

```javascript
const map = L.map('map').setView([20, 0], 2);

L.tileLayer('http://localhost:8080/{z}/{x}/{y}.png', {
    attribution: 'xyztiles',
    tileSize: 256, 
    // 256 makes 512 tiles crisp on highres/retina, but 
    // you can also use 512 and adjust the zoom level with an offset
    maxNativeZoom: 6,
    maxZoom: 10
}).addTo(map);
```

## Interactive Viewer Features

The embedded Leaflet viewer includes:

- **🔍 Debug Mode** - Press `D` key or click the debug button to show tile coordinates and boundaries
- **📏 Scale Control** - Imperial and metric measurements
- **📍 Pan & Zoom** - Standard map navigation
- **ℹ️ Info Panel** - Server statistics and endpoint details
- **🖥️ Console Logging** - Tile load events and coordinate tracking

## How It Works

### Architecture

```
┌─────────────────────────────────────────────────┐
│  Single Go Binary (13MB)                        │
│                                                 │
│  ┌─────────────┐    ┌──────────────────────┐   │
│  │ Embedded    │───▶│ Tile Generator       │   │
│  │ world.jpg   │    │ (on-demand + cached) │   │
│  │ (1.6MB)     │    │                      │   │
│  └─────────────┘    └──────────────────────┘   │
│         ▲                    │                  │
│         │                    ▼                  │
│  ┌─────────────┐    ┌──────────────────────┐   │
│  │ --image     │    │ HTTP Server          │   │
│  │ flag        │    │ /{z}/{x}/{y}.png     │   │
│  │ (optional)  │    │ / (Leaflet viewer)   │   │
│  └─────────────┘    └──────────────────────┘   │
│                                                 │
└─────────────────────────────────────────────────┘
```

### Tile Generation Pipeline

1. **XYZ → Geographic Bounds** - Convert tile coordinates to lat/lon using Web Mercator formulas
2. **Geographic → Pixel Bounds** - Map lat/lon to pixel coordinates in the equirectangular source
3. **Extract Region** - Pull the relevant section from the source image
4. **Resample** - Scale to 512×512 using CatmullRom interpolation for high quality
5. **Encode** - Convert to PNG and serve with cache headers

### Coordinate Systems

- **Input Image**: Equirectangular (EPSG:4326) - simple linear mapping
- **Output Tiles**: Web Mercator (EPSG:3857) - standard for web maps
- **Latitude Range**: ±85.0511° (Web Mercator limit)

## Project Structure

```
xyztiles/
├── cmd/               # CLI commands (Cobra)
├── src/
│   ├── imagery/       # Image loading and tile extraction
│   ├── resources/     # Embedded assets (map + viewer HTML)
│   ├── server/        # HTTP server and handlers
│   └── tilemath/      # XYZ coordinate conversions
├── res/               # Source resources (not in binary)
└── main.go            # Entry point
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Verbose output
go test -v ./src/...
```

**Current Test Coverage:**
- `src/imagery`: 100.0%
- `src/resources`: 100.0%
- `src/server`: 87.9%
- `src/tilemath`: 97.7%

### Building

```bash
# Standard build
go build -o xyztiles main.go

# Build for all platforms with GoReleaser
goreleaser build --snapshot --clean

# Release build
goreleaser release --clean
```

### Code Organization

The project follows clean architecture principles:

- **`tilemath`** - Pure coordinate transformation logic
- **`imagery`** - Image loading and tile generation
- **`server`** - HTTP handlers and routing
- **`resources`** - Embedded assets using `//go:embed`

## Use Cases

- **Education** - Teach students about web mapping without external dependencies
- **Offline Demos** - Present web mapping concepts without internet
- **Air-Gapped Environments** - Deploy maps in secure/isolated networks
- **Custom Cartography** - Serve your own custom world map designs
- **Development** - Test mapping applications locally
- **Prototyping** - Quickly spin up a tile server for proof-of-concepts

## Technical Details

### Dependencies

**Runtime:** Zero! Single static binary.

**Build-time:**
- `golang.org/x/image/draw` - Image resampling
- `github.com/spf13/cobra` - CLI framework

### Performance

- **Startup**: ~100ms (loading embedded 1.6MB JPEG)
- **Tile Generation**: ~10-200ms depending on zoom level and complexity
- **Memory**: ~50MB base + loaded image
- **Concurrency**: Handles multiple simultaneous tile requests

### Limitations

- **Max Zoom**: Native tiles only go to zoom 6 (higher zooms are browser-scaled)
- **Projection**: Only equirectangular input images supported currently
- **Format**: JPEG input only (PNG/TIFF support planned)
- **Caching**: In-memory LRU cache not yet implemented (coming soon)

## Roadmap

- [ ] PNG and TIFF input support
- [ ] CORS configuration
- [ ] Tile export to disk (directory, MBTiles. PMTiles)
- [ ] Docker image
- [ ] Prometheus metrics endpoint
- [ ] In-memory LRU tile cache
- [ ] Pre-warm cache at startup option
- [ ] Multiple embedded base maps


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Credits

- **Map Data**: NASA Blue Marble (https://visibleearth.nasa.gov/)
- **Mapping Library**: Leaflet (https://leafletjs.com/)
- **Tile Math**: Based on OSM Slippy Map standard (https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames)

## Acknowledgments

Built as an educational tool to help newcomers understand web mapping without the complexity of external tile services. Complements vector tile servers and helps bridge the gap between static maps and interactive web cartography.

---

**Made with ❤️ for the mapping community**
