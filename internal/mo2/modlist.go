package mo2

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// ModlistEntry is one line from modlist.txt (+/- prefix).
type ModlistEntry struct {
	Name    string
	Enabled bool
	Order   int
}

// ParseModlist reads modlist.txt: lines are "+Name" or "-Name"; order is 0-based in file order.
func ParseModlist(path string) ([]ModlistEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read modlist: %w", err)
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	var out []ModlistEntry
	sc := bufio.NewScanner(bytes.NewReader(raw))
	order := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) < 2 {
			continue
		}
		prefix, name := line[0], strings.TrimSpace(line[1:])
		if name == "" {
			continue
		}
		if !utf8.ValidString(name) {
			continue
		}
		var enabled bool
		switch prefix {
		case '+':
			enabled = true
		case '-':
			enabled = false
		default:
			continue
		}
		out = append(out, ModlistEntry{Name: name, Enabled: enabled, Order: order})
		order++
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan modlist: %w", err)
	}
	return out, nil
}
