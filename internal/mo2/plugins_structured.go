package mo2

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PluginEntry is one plugin line from plugins.txt (MO2: leading * often marks inactive).
type PluginEntry struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// ParsePluginsFile reads plugins.txt and returns structured entries (skips empty/comments).
func ParsePluginsFile(profileDir string) ([]PluginEntry, error) {
	p := filepath.Join(profileDir, "plugins.txt")
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read plugins.txt: %w", err)
	}
	lines := parsePluginLinesFromBytes(raw)
	var out []PluginEntry
	for _, line := range lines {
		active := true
		s := line
		if strings.HasPrefix(s, "*") {
			active = false
			s = strings.TrimSpace(s[1:])
		}
		if s == "" {
			continue
		}
		out = append(out, PluginEntry{Name: s, Active: active})
	}
	return out, nil
}

func parsePluginLinesFromBytes(raw []byte) []string {
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
