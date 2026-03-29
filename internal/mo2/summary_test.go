package mo2

import (
	"path/filepath"
	"testing"
)

func TestBuildProfileSummary(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	s, err := BuildProfileSummary(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if s.EnabledMods != 4 || s.DisabledMods != 1 || s.TotalModlistEntries != 5 {
		t.Errorf("counts: %+v", s)
	}
	if s.DuplicateModNames != 1 {
		t.Errorf("duplicate mod names: %d", s.DuplicateModNames)
	}
	if !s.HasPluginsTxt || s.PluginsLineCount != 2 {
		t.Errorf("plugins: %+v", s)
	}
	if !s.HasLoadorderTxt || s.LoadorderLineCount != 1 {
		t.Errorf("loadorder: %+v", s)
	}
	if s.ModsMissingFolder != 2 {
		t.Errorf("missing folder: %d (GhostMod + DisabledMod)", s.ModsMissingFolder)
	}
}
