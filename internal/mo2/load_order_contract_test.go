package mo2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPluginsOrdered_fallback(t *testing.T) {
	dir := t.TempDir()
	prof := filepath.Join(dir, "prof")
	if err := mkdirWrite(prof, "plugins.txt", "# p\nA.esp\n*B.esp\n"); err != nil {
		t.Fatal(err)
	}
	ents, w := BuildPluginsOrdered(prof)
	if len(ents) != 2 {
		t.Fatalf("got %+v", ents)
	}
	if ents[0].Name != "A.esp" || !ents[0].Active || ents[1].Name != "B.esp" || ents[1].Active {
		t.Fatalf("entries %+v", ents)
	}
	var sawFallback bool
	for _, x := range w {
		if x == "load_order_source=plugins_txt_fallback" {
			sawFallback = true
		}
	}
	if !sawFallback {
		t.Fatalf("warnings %v", w)
	}
}

func mkdirWrite(dir, name, content string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}
