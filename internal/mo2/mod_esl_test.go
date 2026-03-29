package mo2

import (
	"path/filepath"
	"testing"
)

func TestListModPluginArchives(t *testing.T) {
	mods := filepath.Join("testdata", "mods")
	rel, w, err := ListModPluginArchives(mods, "DeepPlugins", 10, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(rel) != 1 || rel[0] != "DeepPlugins/sub/deep/test.esp" {
		t.Fatalf("rel %v warn %v", rel, w)
	}
}

func TestListModPluginArchives_maxFiles(t *testing.T) {
	mods := filepath.Join("testdata", "mods")
	_, w, err := ListModPluginArchives(mods, "DeepPlugins", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	// maxFiles defaults to 200; with 1 file no truncation
	if len(w) != 0 {
		t.Fatalf("unexpected warn: %v", w)
	}
}

func TestListModPluginArchives_invalidName(t *testing.T) {
	_, _, err := ListModPluginArchives(filepath.Join("testdata", "mods"), "../x", 5, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}
