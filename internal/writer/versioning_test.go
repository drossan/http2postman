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

func TestOutputPath(t *testing.T) {
	got := OutputPath("output", "Griddo API")
	if got != "output/griddo_api.json" {
		t.Errorf("got %q, want %q", got, "output/griddo_api.json")
	}
}

func TestReadExistingVersion_NoFile(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	_, _, _, found := ReadExistingVersion(memFS, "nonexistent.json")
	if found {
		t.Error("expected found=false for nonexistent file")
	}
}

func TestReadExistingVersion_WithVersion(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("col.json", []byte(`{"info":{"name":"Test","version":"1.2.3"}}`), 0644)

	major, minor, patch, found := ReadExistingVersion(memFS, "col.json")
	if !found {
		t.Fatal("expected found=true")
	}
	if major != 1 || minor != 2 || patch != 3 {
		t.Errorf("expected 1.2.3, got %d.%d.%d", major, minor, patch)
	}
}

func TestReadExistingVersion_NoVersionField(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("col.json", []byte(`{"info":{"name":"Test"}}`), 0644)

	_, _, _, found := ReadExistingVersion(memFS, "col.json")
	if found {
		t.Error("expected found=false when version field is missing")
	}
}

func TestBumpVersion_Minor(t *testing.T) {
	v := BumpVersion(1, 2, 0, BumpMinor)
	if v != "1.3.0" {
		t.Errorf("got %q, want %q", v, "1.3.0")
	}
}

func TestBumpVersion_Patch(t *testing.T) {
	v := BumpVersion(1, 2, 3, BumpPatch)
	if v != "1.2.4" {
		t.Errorf("got %q, want %q", v, "1.2.4")
	}
}

func TestBumpVersion_Major(t *testing.T) {
	v := BumpVersion(1, 2, 3, BumpMajor)
	if v != "2.0.0" {
		t.Errorf("got %q, want %q", v, "2.0.0")
	}
}

func TestResolveVersionedOutput_NoExisting(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpMinor)
	if path != "output/griddo_api.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api.json")
	}
	if version != "1.0.0" {
		t.Errorf("version: got %q, want %q", version, "1.0.0")
	}
}

func TestResolveVersionedOutput_ExistingMinorBump(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api.json", []byte(`{"info":{"version":"1.0.0"}}`), 0644)

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpMinor)
	if path != "output/griddo_api.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api.json")
	}
	if version != "1.1.0" {
		t.Errorf("version: got %q, want %q", version, "1.1.0")
	}
}

func TestResolveVersionedOutput_ExistingPatchBump(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api.json", []byte(`{"info":{"version":"1.1.0"}}`), 0644)

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpPatch)
	if path != "output/griddo_api.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api.json")
	}
	if version != "1.1.1" {
		t.Errorf("version: got %q, want %q", version, "1.1.1")
	}
}

func TestResolveVersionedOutput_ExistingMajorBump(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output/griddo_api.json", []byte(`{"info":{"version":"1.2.3"}}`), 0644)

	path, version := ResolveVersionedOutput(memFS, "output", "Griddo API", BumpMajor)
	if path != "output/griddo_api.json" {
		t.Errorf("path: got %q, want %q", path, "output/griddo_api.json")
	}
	if version != "2.0.0" {
		t.Errorf("version: got %q, want %q", version, "2.0.0")
	}
}
