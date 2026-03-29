package mo2

import (
	"path/filepath"
	"testing"
)

func TestParseModlist(t *testing.T) {
	path := filepath.Join("testdata", "profile", "modlist.txt")
	entries, err := ParseModlist(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("got %d entries, want 5", len(entries))
	}
	cases := []struct {
		i       int
		name    string
		enabled bool
	}{
		{0, "SkyUI", true},
		{1, "DisabledMod", false},
		{2, "NoMetaMod", true},
		{3, "GhostMod", true},
		{4, "SkyUI", true},
	}
	for _, c := range cases {
		e := entries[c.i]
		if e.Name != c.name || e.Enabled != c.enabled || e.Order != c.i {
			t.Errorf("[%d] got %+v, want name=%q enabled=%v order=%d", c.i, e, c.name, c.enabled, c.i)
		}
	}
}
