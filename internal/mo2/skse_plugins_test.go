package mo2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSKSEPluginInventory_winnerByModlistOrder(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_skse"),
		ModsDir:    filepath.Join("testdata", "mods_skse"),
	}
	opts := DefaultSKSEPluginOptions()
	opts.MaxDLLs = 100
	report, err := BuildSKSEPluginInventory(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	var pluginRow *SKSEPluginRow
	for i := range report.Plugins {
		if report.Plugins[i].VirtualPath == "SKSE/Plugins/plugin.dll" {
			pluginRow = &report.Plugins[i]
			break
		}
	}
	if pluginRow == nil {
		t.Fatalf("expected SKSE/Plugins/plugin.dll in plugins: %#v", report.Plugins)
	}
	if pluginRow.Winner.Name != "ModHigh" {
		t.Errorf("winner: got %q want ModHigh", pluginRow.Winner.Name)
	}
	if len(pluginRow.Providers) != 2 {
		t.Errorf("providers: got %d want 2", len(pluginRow.Providers))
	}
}

func TestBuildSKSEPluginInventory_skseRootWithoutDataPrefix(t *testing.T) {
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_skse"),
		ModsDir:    filepath.Join("testdata", "mods_skse"),
	}
	report, err := BuildSKSEPluginInventory(cfg, DefaultSKSEPluginOptions())
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, row := range report.Plugins {
		if row.VirtualPath == "SKSE/Plugins/from_skse_root.dll" {
			found = true
			if row.Winner.Name != "ModHigh" {
				t.Errorf("winner: %q", row.Winner.Name)
			}
		}
	}
	if !found {
		t.Fatal("missing SKSE/Plugins/from_skse_root.dll")
	}
}

func TestBuildSKSEPluginInventory_gameDataLowestPriority(t *testing.T) {
	tmp := t.TempDir()
	gamePlugins := filepath.Join(tmp, "SKSE", "Plugins")
	if err := os.MkdirAll(gamePlugins, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gamePlugins, "shared.dll"), []byte("game"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		ProfileDir: filepath.Join("testdata", "profile_skse"),
		ModsDir:    filepath.Join("testdata", "mods_skse"),
	}
	opts := DefaultSKSEPluginOptions()
	opts.GameDataDir = tmp
	report, err := BuildSKSEPluginInventory(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	var row *SKSEPluginRow
	for i := range report.Plugins {
		if report.Plugins[i].VirtualPath == "SKSE/Plugins/shared.dll" {
			row = &report.Plugins[i]
			break
		}
	}
	if row == nil {
		t.Fatal("expected shared.dll from game + mod overlay")
	}
	if row.Winner.Name != "ModHigh" {
		t.Fatalf("mod should override game: winner=%+v providers=%+v", row.Winner, row.Providers)
	}
	names := make([]string, len(row.Providers))
	for i, p := range row.Providers {
		names[i] = p.Name
	}
	if len(names) < 2 {
		t.Fatalf("expected game + mods providers: %v", names)
	}
}

func TestBuildSKSEPluginInventory_maxDllsTruncation(t *testing.T) {
	tmp := t.TempDir()
	prof := filepath.Join(tmp, "profile")
	mods := filepath.Join(tmp, "mods")
	if err := os.MkdirAll(filepath.Join(prof), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prof, "modlist.txt"), []byte("+OneMod\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	modRoot := filepath.Join(mods, "OneMod", "Data", "SKSE", "Plugins")
	if err := os.MkdirAll(modRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		name := filepath.Join(modRoot, string(rune('a'+i))+".dll")
		if err := os.WriteFile(name, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := Config{ProfileDir: prof, ModsDir: mods}
	opts := DefaultSKSEPluginOptions()
	opts.MaxDLLs = 3
	report, err := BuildSKSEPluginInventory(cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Truncated {
		t.Error("expected Truncated true")
	}
	if len(report.Plugins) != 3 {
		t.Errorf("plugins len: got %d want 3", len(report.Plugins))
	}
}
