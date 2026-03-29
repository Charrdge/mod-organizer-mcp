package mo2

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// ParseMetaINI reads MO2 meta.ini into a flat map "section.key" -> value (trimmed).
// Empty lines and # / ; comments are skipped.
func ParseMetaINI(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	return parseINIBytes(raw)
}

func parseINIBytes(raw []byte) (map[string]string, error) {
	out := make(map[string]string)
	section := "General"
	sc := bufio.NewScanner(bytes.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			if section == "" {
				section = "General"
			}
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		k := section + "." + key
		out[k] = val
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan ini: %w", err)
	}
	return out, nil
}
