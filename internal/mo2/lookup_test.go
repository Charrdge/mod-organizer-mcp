package mo2

import (
	"path/filepath"
	"testing"
)

func TestLookupMod_exact(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	res, err := LookupMod(cfg, "SkyUI")
	if err != nil {
		t.Fatal(err)
	}
	if res.Match == nil || res.Match.Name != "SkyUI" || !res.Match.Enabled {
		t.Fatalf("match: %+v", res)
	}
	if res.Match.Meta["General.modid"] != "2014" {
		t.Fatalf("meta: %+v", res.Match.Meta)
	}
}

func TestLookupMod_disabled(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	res, err := LookupMod(cfg, "DisabledMod")
	if err != nil {
		t.Fatal(err)
	}
	if res.Match == nil || res.Match.Enabled {
		t.Fatalf("match: %+v", res.Match)
	}
}

func TestLookupMod_notFound(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	res, err := LookupMod(cfg, "NopeMod")
	if err != nil {
		t.Fatal(err)
	}
	if !res.NotFound {
		t.Fatalf("expected not_found: %+v", res)
	}
}

func TestLookupMod_prefix(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile"),
		ModsDir:    filepath.Join("testdata", "mods"),
	}
	res, err := LookupMod(cfg, "NoMe")
	if err != nil {
		t.Fatal(err)
	}
	if res.Match == nil || res.Match.Name != "NoMetaMod" {
		t.Fatalf("prefix match: %+v", res)
	}
}
