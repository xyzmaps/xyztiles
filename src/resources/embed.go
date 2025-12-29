package resources

import (
	_ "embed"
)

// DefaultWorldMap contains the embedded Blue Marble world map image
// This is embedded at compile time from the res/ directory
//
//go:embed world.topo.200407.3x5400x2700.jpg
var DefaultWorldMap []byte

// HasEmbeddedMap returns true if the default world map is embedded
func HasEmbeddedMap() bool {
	return len(DefaultWorldMap) > 0
}

// DefaultMapSize returns the size of the embedded map in bytes
func DefaultMapSize() int {
	return len(DefaultWorldMap)
}
