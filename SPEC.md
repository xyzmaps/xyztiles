# xyztiles: Embedded World Map Tile Server

## Project Goal

A single Go binary that serves web map tiles from an embedded or user-provided world map image. Zero external dependencies, no hosted services required - perfect for learning web mapping or offline/air-gapped environments.

## Core Requirements

### Input Constraints (Simplified Scope)
- Source image: EPSG:4326 (equirectangular) projection
- Bounding box: Full world extent (-180, -90, 180, 90)
- Format: TIFF or JPEG (common image formats, no exotic GIS formats)
- Default: Embed a downscaled NASA Blue Marble (~5MB compressed)

### Output
- Standard XYZ tile scheme (same as OpenStreetMap, Leaflet default)
- Tile size: 512×512 PNG
- Zoom levels: 0-6 (higher zooms just scale up z6 tiles in browser)
- URL pattern: `/{z}/{x}/{y}.png`

### Architecture
```
┌─────────────────────────────────────────────────┐
│  Single Go Binary                               │
│                                                 │
│  ┌─────────────┐    ┌──────────────────────┐   │
│  │ Embedded    │───▶│ Tile Generator       │   │
│  │ world.png   │    │ (on-demand + cached) │   │
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

## Technical Design

### 1. Tile Math (XYZ ↔ Geographic Coordinates)

Standard Web Mercator tile formulas:
- Number of tiles at zoom z: `2^z × 2^z`
- Tile (x, y, z) → Mercator bounds → Lat/Lon bounds
- Need to handle Mercator latitude clamping (~±85.051°)

Reference: https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames

### 2. Coordinate Transformation Pipeline

```
Tile (z, x, y)
    ↓
Mercator bounds (meters)
    ↓
Lat/Lon bounds (degrees)  
    ↓
Source pixel bounds (equirectangular is linear!)
    ↓
Sample/resample to 512×512 tile
    ↓
Encode as PNG
```

For equirectangular source spanning -180 to 180, -90 to 90:
- `pixel_x = (lon + 180) / 360 * image_width`
- `pixel_y = (90 - lat) / 180 * image_height`

### 3. Resampling Strategy

- Use bilinear interpolation (good balance of speed/quality)
- Go's `golang.org/x/image/draw` has `draw.BiLinear` 
- At low zoom (0-2), source region is large → downsampling
- At high zoom (5-6), source region is small → upsampling

### 4. Caching

- In-memory LRU cache for generated tiles
- Cache key: `z/x/y` 
- Configurable size (default: ~100MB worth of tiles)
- Optional: flag to pre-warm cache at startup

### 5. HTTP Server

Endpoints:
- `GET /{z}/{x}/{y}.png` - tile endpoint
- `GET /` - embedded Leaflet viewer HTML

Headers:
- `Cache-Control: max-age=86400` (tiles are immutable)
- `Content-Type: image/png`

### 6. CLI Interface

```bash
# Run with embedded default map
./xyztiles

# Run with custom image
./xyztiles --image path/to/world.tif

# Custom port
./xyztiles --port 8080

# Pre-generate all tiles to disk (optional feature)
./xyztiles --export ./tiles
```

## Implementation Phases

### Phase 1: Core Tile Generation
- [ ] Tile coordinate math (z/x/y → lat/lon bounds)
- [ ] Load equirectangular image (JPEG/TIFF)
- [ ] Extract region for tile bounds
- [ ] Resample to 512×512 with bilinear interpolation
- [ ] Encode as PNG
- [ ] Unit tests with known tile coordinates

### Phase 2: HTTP Server
- [ ] Basic tile endpoint `/{z}/{x}/{y}.png`
- [ ] Embedded Leaflet viewer at `/`
- [ ] Proper error handling (404 for invalid tiles)
- [ ] CORS headers for local development

### Phase 3: Caching & Performance  
- [ ] In-memory LRU cache
- [ ] Cache size configuration
- [ ] Optional cache pre-warming
- [ ] Concurrent request handling (tiles generated once)

### Phase 4: Polish
- [ ] Embed default Blue Marble image using `//go:embed`
- [ ] CLI flags (--port, --image, --cache-size)
- [ ] Graceful shutdown
- [ ] Optional: --export flag to write tiles to disk

## Dependencies

Minimal, all well-maintained:
- `golang.org/x/image/draw` - resampling
- `golang.org/x/image/tiff` - TIFF reading (if supporting TIFF input)
- Standard library for everything else (net/http, image/png, embed)

## Testing Strategy

- Unit tests for tile math (known z/x/y → known bounds)
- Golden tests: generate specific tiles, compare to expected output
- Integration test: spin up server, fetch tiles via HTTP

## Reference Resources

- [OSM Slippy Map Tilenames](https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames) - tile math formulas
- [Bing Maps Tile System](https://docs.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system) - detailed explanation
- [NASA Blue Marble](https://visibleearth.nasa.gov/collection/1484/blue-marble) - source imagery
- [Leaflet Quick Start](https://leafletjs.com/examples/quick-start/) - for the embedded viewer

## Notes from Discussion

- User has existing workflow using `gdal2tiles.py` with `--zoom=0-6 --xyz --tilesize=512`
- Goal is educational: help newcomers to web mapping get started easily  
- Will complement a separate MVT-OSM vector tile server project
- Pure Go preferred over CGO bindings to GDAL for portability