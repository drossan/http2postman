package writer

import (
	"testing"

	"github.com/drossan/http2postman/internal/fs"
)

func TestNameToSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Griddo API", "griddo_api"},
		{"My Cool Collection", "my_cool_collection"},
		{"already_slug", "already_slug"},
		{"  Spaces  Around  ", "spaces_around"},
		{"Mixed-Dash_Under", "mixed_dash_under"},
		{"UPPER CASE", "upper_case"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NameToSlug(tt.input)
			if got != tt.expected {
				t.Errorf("NameToSlug(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFindLatestVersion_NoExisting(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.MkdirAll("output", 0755)

	major, minor, patch, found := FindLatestVersion(memFS, "output", "Griddo API")
	if found {
		t.Error("expected found=false when no files exist")
	}
	if major != 0 || minor != 0 || patch != 0 {
		t.Errorf("expected 0.0.0, got %d.%d.%d", major, minor, patch)
	}
}

func TestFindLatestVersion_SingleVersion(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api_1_0_0.json", []byte("{}"), 0644)

	major, minor, patch, found := FindLatestVersion(memFS, "output", "Griddo API")
	if !found {
		t.Error("expected found=true")
	}
	if major != 1 || minor != 0 || patch != 0 {
		t.Errorf("expected 1.0.0, got %d.%d.%d", major, minor, patch)
	}
}

func TestFindLatestVersion_MultipleVersions(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api_1_0_0.json", []byte("{}"), 0644)
	memFS.WriteFile("output/griddo_api_1_1_0.json", []byte("{}"), 0644)
	memFS.WriteFile("output/griddo_api_1_2_0.json", []byte("{}"), 0644)

	major, minor, patch, found := FindLatestVersion(memFS, "output", "Griddo API")
	if !found {
		t.Error("expected found=true")
	}
	if major != 1 || minor != 2 || patch != 0 {
		t.Errorf("expected 1.2.0, got %d.%d.%d", major, minor, patch)
	}
}

func TestBuildVersionedPath_Minor(t *testing.T) {
	path, version := BuildVersionedPath("output", "griddo_api", 1, 2, 0, BumpMinor)
	if path != "output/griddo_api_1_3_0.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api_1_3_0.json")
	}
	if version != "1.3.0" {
		t.Errorf("version: got %q, want %q", version, "1.3.0")
	}
}

func TestBuildVersionedPath_Patch(t *testing.T) {
	path, version := BuildVersionedPath("output", "griddo_api", 1, 2, 3, BumpPatch)
	if path != "output/griddo_api_1_2_4.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api_1_2_4.json")
	}
	if version != "1.2.4" {
		t.Errorf("version: got %q, want %q", version, "1.2.4")
	}
}

func TestBuildVersionedPath_Major(t *testing.T) {
	path, version := BuildVersionedPath("output", "griddo_api", 1, 2, 3, BumpMajor)
	if path != "output/griddo_api_2_0_0.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api_2_0_0.json")
	}
	if version != "2.0.0" {
		t.Errorf("version: got %q, want %q", version, "2.0.0")
	}
}

func TestResolveVersionedOutput_NoExisting(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.MkdirAll("output", 0755)

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpMinor)
	if path != "output/griddo_api_1_0_0.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api_1_0_0.json")
	}
	if version != "1.0.0" {
		t.Errorf("version: got %q, want %q", version, "1.0.0")
	}
}

func TestResolveVersionedOutput_ExistingMinorBump(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api_1_0_0.json", []byte("{}"), 0644)

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpMinor)
	if path != "output/griddo_api_1_1_0.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api_1_1_0.json")
	}
	if version != "1.1.0" {
		t.Errorf("version: got %q, want %q", version, "1.1.0")
	}
}

func TestResolveVersionedOutput_ExistingPatchBump(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api_1_0_0.json", []byte("{}"), 0644)

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpPatch)
	if path != "output/griddo_api_1_0_1.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api_1_0_1.json")
	}
	if version != "1.0.1" {
		t.Errorf("version: got %q, want %q", version, "1.0.1")
	}
}

func TestResolveVersionedOutput_ExistingMajorBump(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api_1_2_3.json", []byte("{}"), 0644)

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpMajor)
	if path != "output/griddo_api_2_0_0.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api_2_0_0.json")
	}
	if version != "2.0.0" {
		t.Errorf("version: got %q, want %q", version, "2.0.0")
	}
}
