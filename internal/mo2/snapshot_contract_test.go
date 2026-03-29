package mo2

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSnapshot_contract(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	snap, err := BuildSnapshot(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if snap.SnapshotContractVersion != "1" {
		t.Fatalf("contract version: %q", snap.SnapshotContractVersion)
	}
	if snap.ProfileListPaths == nil {
		t.Fatal("expected profile_list_paths")
	}
	if !snap.ProfileListPaths.PluginsTxt.Present || !strings.HasSuffix(snap.ProfileListPaths.PluginsTxt.Path, "plugins.txt") {
		t.Fatalf("plugins_txt: %+v", snap.ProfileListPaths.PluginsTxt)
	}
	if !snap.ProfileListPaths.LoadorderTxt.Present {
		t.Fatalf("loadorder_txt: %+v", snap.ProfileListPaths.LoadorderTxt)
	}
	if len(snap.ProfileIni) != len(profileIniWhitelist) {
		t.Fatalf("profile_ini len want %d got %d", len(profileIniWhitelist), len(snap.ProfileIni))
	}
	for _, e := range snap.ProfileIni {
		if e.Present {
			t.Errorf("unexpected present ini in testdata: %s", e.Basename)
		}
	}
	if len(snap.ArchiveSearchRoots) == 0 {
		t.Fatal("expected archive_search_roots")
	}
	if len(snap.PluginsOrdered) != 0 {
		t.Fatalf("plugins_ordered should be off by default, got %d", len(snap.PluginsOrdered))
	}
}

func TestBuildSnapshot_includeContractFalse(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	opts := DefaultSnapshotOptions()
	opts.IncludeContract = false
	snap, err := BuildSnapshotWithOptions(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if snap.SnapshotContractVersion != "" || snap.ProfileListPaths != nil || len(snap.ProfileIni) != 0 || len(snap.ArchiveSearchRoots) != 0 {
		t.Fatalf("contract should be empty: %+v", snap)
	}
}

func TestBuildMachineContractPayload(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	p, err := BuildMachineContractPayload(cfg, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if p.SnapshotContractVersion != "1" || len(p.ProfileIni) != len(profileIniWhitelist) {
		t.Fatalf("unexpected payload: version=%q ini=%d", p.SnapshotContractVersion, len(p.ProfileIni))
	}
	if !p.ProfileListPaths.PluginsTxt.Present {
		t.Fatal("plugins.txt should exist")
	}
	if len(p.ArchiveSearchRoots) == 0 {
		t.Fatal("expected roots")
	}
}

func TestBuildPluginLoadOrderPayload(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	p, err := BuildPluginLoadOrderPayload(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.PluginsOrdered) != 1 || p.PluginsOrdered[0].Name != "Skyrim.esm" {
		t.Fatalf("%+v", p.PluginsOrdered)
	}
	if !p.ProfileListPaths.PluginsTxt.Present {
		t.Fatal("expected plugins path present")
	}
}

func TestBuildSnapshot_includePluginLoadOrder(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	opts := DefaultSnapshotOptions()
	opts.IncludePluginLoadOrder = true
	snap, err := BuildSnapshotWithOptions(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.PluginsOrdered) != 1 || snap.PluginsOrdered[0].Name != "Skyrim.esm" || !snap.PluginsOrdered[0].Active {
		t.Fatalf("plugins_ordered: %+v", snap.PluginsOrdered)
	}
	var hasMissing bool
	for _, w := range snap.Warnings {
		if strings.Contains(w, "plugin_in_plugins_txt_missing_from_load_order:Disabled.esp") {
			hasMissing = true
		}
	}
	if !hasMissing {
		t.Errorf("expected warning about Disabled.esp missing from load order, warnings=%v", snap.Warnings)
	}
}
