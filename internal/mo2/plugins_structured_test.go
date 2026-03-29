package mo2

import (
	"path/filepath"
	"testing"
)

func TestParsePluginsFile(t *testing.T) {
	ents, err := ParsePluginsFile(filepath.Join("testdata", "profile"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 2 {
		t.Fatalf("len %d %+v", len(ents), ents)
	}
	if ents[0].Name != "Skyrim.esm" || !ents[0].Active {
		t.Fatalf("first: %+v", ents[0])
	}
	if ents[1].Name != "Disabled.esp" || ents[1].Active {
		t.Fatalf("second: %+v", ents[1])
	}
}
