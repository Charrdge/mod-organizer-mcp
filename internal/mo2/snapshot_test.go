package mo2

import (
	"path/filepath"
	"testing"
)

func TestBuildSnapshot(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	snap, err := BuildSnapshot(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.PluginLines) != 2 {
		t.Errorf("plugin_lines: %v", snap.PluginLines)
	}
	if len(snap.LoadorderLines) != 1 || snap.LoadorderLines[0] != "Skyrim.esm" {
		t.Errorf("loadorder_lines: %v", snap.LoadorderLines)
	}
	if len(snap.Mods) != 5 {
		t.Fatalf("mods len %d", len(snap.Mods))
	}
	sky := snap.Mods[0]
	if sky.Name != "SkyUI" || !sky.Enabled || sky.Meta["General.modid"] != "2014" {
		t.Errorf("first row: %+v", sky)
	}
	noMeta := snap.Mods[2]
	if len(noMeta.Warnings) == 0 || noMeta.Warnings[0] != "meta.ini missing" {
		t.Errorf("NoMetaMod warnings: %v", noMeta.Warnings)
	}
	ghost := snap.Mods[3]
	if len(ghost.Warnings) == 0 || ghost.Warnings[0] != "mod folder missing under MO2_MODS_DIR" {
		t.Errorf("GhostMod warnings: %v", ghost.Warnings)
	}
	dup := snap.Mods[4]
	if dup.Name != "SkyUI" {
		t.Fatal("expected duplicate SkyUI")
	}
	dupOk := false
	for _, w := range dup.Warnings {
		if w == "duplicate name earlier in modlist.txt" {
			dupOk = true
		}
	}
	if !dupOk {
		t.Errorf("duplicate SkyUI should warn: %v", dup.Warnings)
	}
}
