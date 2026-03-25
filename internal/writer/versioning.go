package writer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/drossan/http2postman/internal/fs"
)

// BumpType represents the type of semantic version bump.
type BumpType int

const (
	BumpPatch BumpType = iota
	BumpMinor
	BumpMajor
)

var multiSpaceRegexp = regexp.MustCompile(`\s+`)

// NameToSlug converts a collection name to a filename-safe slug.
// "Griddo API" → "griddo_api"
func NameToSlug(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = multiSpaceRegexp.ReplaceAllString(name, "_")
	return name
}

// OutputPath returns the canonical output file path for a collection name.
// The file is always the same: "griddo_api.json" for "Griddo API".
func OutputPath(dir string, collectionName string) string {
	return filepath.Join(dir, NameToSlug(collectionName)+".json")
}

// ReadExistingVersion reads the version from an existing collection JSON file.
// Returns 0,0,0,false if the file doesn't exist or has no version.
func ReadExistingVersion(fsys fs.FileSystem, path string) (major, minor, patch int, found bool) {
	data, err := fsys.ReadFile(path)
	if err != nil {
		return 0, 0, 0, false
	}

	var col struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &col); err != nil || col.Info.Version == "" {
		return 0, 0, 0, false
	}

	return parseVersion(col.Info.Version)
}

func parseVersion(v string) (major, minor, patch int, ok bool) {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	ma, err1 := strconv.Atoi(parts[0])
	mi, err2 := strconv.Atoi(parts[1])
	pa, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return ma, mi, pa, true
}

// BumpVersion applies the given bump type and returns the new version string.
func BumpVersion(major, minor, patch int, bump BumpType) string {
	switch bump {
	case BumpMajor:
		major++
		minor = 0
		patch = 0
	case BumpMinor:
		minor++
		patch = 0
	case BumpPatch:
		patch++
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// ResolveVersionedOutput determines the output path and version for a collection.
// If no previous file exists, returns version 1.0.0.
// If a previous file exists, applies the given bump to its version.
func ResolveVersionedOutput(fsys fs.FileSystem, dir string, collectionName string, bump BumpType) (string, string) {
	path := OutputPath(dir, collectionName)
	major, minor, patch, found := ReadExistingVersion(fsys, path)

	if !found {
		return path, "1.0.0"
	}

	return path, BumpVersion(major, minor, patch, bump)
}
