package mo2

import (
	"path/filepath"
	"testing"
)

func TestBuildSnapshotWithOptions_noMeta(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	opts := DefaultSnapshotOptions()
	opts.IncludeMeta = false
	snap, err := BuildSnapshotWithOptions(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range snap.Mods {
		if len(m.Meta) > 0 {
			t.Errorf("expected no meta for %s", m.Name)
		}
	}
}

func TestBuildSnapshotWithOptions_onlyEnabled(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	opts := DefaultSnapshotOptions()
	opts.OnlyEnabled = true
	snap, err := BuildSnapshotWithOptions(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range snap.Mods {
		if !m.Enabled {
			t.Errorf("disabled mod leaked: %s", m.Name)
		}
	}
	if len(snap.Mods) != 4 {
		t.Fatalf("want 4 enabled, got %d", len(snap.Mods))
	}
}

func TestBuildSnapshotWithOptions_noPluginLines(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	opts := DefaultSnapshotOptions()
	opts.IncludePluginLines = false
	opts.IncludeLoadorderLines = false
	snap, err := BuildSnapshotWithOptions(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.PluginLines) != 0 || len(snap.LoadorderLines) != 0 {
		t.Fatalf("expected no plugin/loadorder lines")
	}
}
