package writer

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

func sampleCollection() *model.PostmanCollection {
	return &model.PostmanCollection{
		Info: model.PostmanInfo{
			Name:   "Test",
			Schema: model.PostmanSchemaV210,
		},
		Item: []model.PostmanItem{
			{
				Name: "Get Users",
				Request: &model.PostmanReq{
					Method: "GET",
					URL:    model.PostmanURL{Raw: "https://api.example.com/users"},
				},
			},
		},
	}
}

func TestPostmanWriter_Write_ValidCollection(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	w := NewPostmanWriter(memFS)

	err := w.Write(sampleCollection(), "output.json", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := memFS.ReadFile("output.json")
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}

	var col model.PostmanCollection
	if err := json.Unmarshal(data, &col); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if col.Info.Name != "Test" {
		t.Errorf("name: got %q", col.Info.Name)
	}
}

func TestPostmanWriter_Write_OverwriteProtection(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output.json", []byte("existing"), 0644)
	w := NewPostmanWriter(memFS)

	err := w.Write(sampleCollection(), "output.json", false)
	if err == nil {
		t.Fatal("expected error for existing file without force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists': %v", err)
	}
}

func TestPostmanWriter_Write_ForceOverwrite(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("output.json", []byte("old"), 0644)
	w := NewPostmanWriter(memFS)

	err := w.Write(sampleCollection(), "output.json", true)
	if err != nil {
		t.Fatalf("unexpected error with force: %v", err)
	}

	data, _ := memFS.ReadFile("output.json")
	if string(data) == "old" {
		t.Error("file was not overwritten")
	}
}

func TestPostmanWriter_Write_IndentedJSON(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	w := NewPostmanWriter(memFS)

	w.Write(sampleCollection(), "output.json", false)
	data, _ := memFS.ReadFile("output.json")

	if !strings.Contains(string(data), "  ") {
		t.Error("expected indented JSON output")
	}
}
