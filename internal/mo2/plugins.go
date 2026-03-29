package mo2

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

// ReadTextLines reads a MO2 list file (plugins.txt / loadorder.txt): non-empty, non-comment lines.
func ReadTextLines(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
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
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
