package writer

import (
	"fmt"
	stdfs "io/fs"
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

// FindLatestVersion scans the directory for existing versioned files
// matching the collection slug and returns the highest version found.
// Returns found=false if no matching files exist.
func FindLatestVersion(fsys fs.FileSystem, dir string, collectionName string) (major, minor, patch int, found bool) {
	slug := NameToSlug(collectionName)
	prefix := slug + "_"
	versionRegexp := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `(\d+)_(\d+)_(\d+)\.json$`)

	fsys.Walk(dir, func(path string, info stdfs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		matches := versionRegexp.FindStringSubmatch(name)
		if matches == nil {
			return nil
		}
		ma, _ := strconv.Atoi(matches[1])
		mi, _ := strconv.Atoi(matches[2])
		pa, _ := strconv.Atoi(matches[3])

		if !found || compareVersions(ma, mi, pa, major, minor, patch) > 0 {
			major, minor, patch = ma, mi, pa
			found = true
		}
		return nil
	})
	return
}

func compareVersions(ma1, mi1, pa1, ma2, mi2, pa2 int) int {
	if ma1 != ma2 {
		return ma1 - ma2
	}
	if mi1 != mi2 {
		return mi1 - mi2
	}
	return pa1 - pa2
}

// BuildVersionedPath creates the output path and version string by applying
// the given bump type to the provided version.
func BuildVersionedPath(dir string, slug string, major, minor, patch int, bump BumpType) (string, string) {
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

	version := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	filename := fmt.Sprintf("%s_%d_%d_%d.json", slug, major, minor, patch)
	return filepath.Join(dir, filename), version
}

// ResolveVersionedOutput finds the latest existing version for the collection
// and applies the given bump. If no previous version exists, returns 1.0.0.
func ResolveVersionedOutput(fsys fs.FileSystem, dir string, collectionName string, bump BumpType) (string, string) {
	slug := NameToSlug(collectionName)
	major, minor, patch, found := FindLatestVersion(fsys, dir, collectionName)

	if !found {
		filename := fmt.Sprintf("%s_1_0_0.json", slug)
		return filepath.Join(dir, filename), "1.0.0"
	}

	return BuildVersionedPath(dir, slug, major, minor, patch, bump)
}
