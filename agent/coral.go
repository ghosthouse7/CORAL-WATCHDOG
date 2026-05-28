package agent

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// RunQueryRaw runs SQL via Coral, returns raw string output (for AI context).
func RunQueryRaw(sql string) (string, error) {
	cmd := exec.Command(".\\coral.exe", "sql", sql)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(errBuf.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// RunQueryRows runs SQL via Coral, returns parsed rows for the dashboard API.
func RunQueryRows(sql string) ([]map[string]string, error) {
	cmd := exec.Command(".\\coral.exe", "sql", sql)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		stderr := strings.TrimSpace(errBuf.String())
		log.Printf("Coral stderr: %s", stderr)
		return nil, fmt.Errorf("%s", stderr)
	}

	raw := out.String()
	for i, line := range strings.Split(raw, "\n") {
		log.Printf("coral[%02d]: %q", i, line)
	}

	return parseCoralTable(raw)
}

// RunQueryString wraps RunQueryRows returning []map[string]interface{}.
func RunQueryString(sql string) ([]map[string]interface{}, error) {
	rows, err := RunQueryRows(sql)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		m := make(map[string]interface{}, len(r))
		for k, v := range r {
			m[k] = v
		}
		out[i] = m
	}
	return out, nil
}

func parseCoralTable(raw string) ([]map[string]string, error) {
	lines := strings.Split(raw, "\n")

	// Detect if this is an empty result — Coral prints "++" or "| | |" with no data
	// Count non-border, non-empty lines that have actual cell content
	var headers []string
	var rows []map[string]string
	borderCount := 0

	for _, line := range lines {
		line = strings.TrimRight(line, "\r \t")
		if line == "" {
			continue
		}

		if isBorderLine(line) {
			borderCount++
			if headers != nil {
				// After the header separator, data rows follow
			}
			continue
		}

		// Must have a cell separator
		if !strings.Contains(line, "|") && !strings.Contains(line, "│") {
			continue
		}

		cells := splitCells(line)
		if len(cells) == 0 {
			continue
		}

		// Check if all cells are empty (blank row like "|    |    |")
		allEmpty := true
		for _, c := range cells {
			if strings.TrimSpace(c) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue // skip blank data rows
		}

		if headers == nil {
			// First non-border row with content = headers
			headers = cells
		} else if borderCount >= 2 {
			// We've seen at least: top border, header, separator border
			// So this is a data row
			row := make(map[string]string, len(headers))
			for i, h := range headers {
				h = strings.TrimSpace(h)
				if h == "" {
					continue
				}
				val := ""
				if i < len(cells) {
					val = strings.TrimSpace(cells[i])
				}
				row[h] = val
			}
			hasVal := false
			for _, v := range row {
				if v != "" {
					hasVal = true
					break
				}
			}
			if hasVal {
				rows = append(rows, row)
			}
		}
	}

	if rows == nil {
		return []map[string]string{}, nil
	}
	return rows, nil
}

func isBorderLine(line string) bool {
	s := strings.TrimSpace(line)
	if s == "" {
		return false
	}
	// Must start and end with + or box char
	if !strings.HasPrefix(s, "+") && !strings.HasPrefix(s, "┌") &&
		!strings.HasPrefix(s, "├") && !strings.HasPrefix(s, "└") {
		return false
	}
	// All chars must be border chars
	for _, r := range s {
		switch r {
		case '┌', '┐', '└', '┘', '├', '┤', '┬', '┴', '┼', '─',
			'+', '-', '=', ' ':
			// ok
		default:
			return false
		}
	}
	return true
}

func splitCells(line string) []string {
	sep := "|"
	if strings.Contains(line, "│") {
		sep = "│"
	}
	parts := strings.Split(line, sep)
	var cells []string
	for _, p := range parts {
		cells = append(cells, strings.TrimSpace(p))
	}
	// Trim outer empty strings from border chars
	for len(cells) > 0 && cells[0] == "" {
		cells = cells[1:]
	}
	for len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
