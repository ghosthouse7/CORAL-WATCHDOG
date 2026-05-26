package agent

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os/exec"
	"strings"
)

// CoralRow represents a single result row from a Coral SQL query.
// Keys are column names, values are string representations.
type CoralRow map[string]string

// RunQuery executes a SQL query via the Coral CLI and returns rows.
// Coral outputs CSV by default, which we parse here.
func RunQuery(sql string) ([]CoralRow, error) {
	// --no-color avoids ANSI codes breaking CSV parse
	cmd := exec.Command(".\\coral.exe", "sql", sql)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("coral sql failed: %w\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []CoralRow{}, nil
	}

	reader := csv.NewReader(strings.NewReader(output))
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse coral output as CSV: %w\nraw output: %s", err, output)
	}

	if len(records) < 2 {
		// Only header row or empty — no data
		return []CoralRow{}, nil
	}

	headers := records[0]
	rows := make([]CoralRow, 0, len(records)-1)

	for _, record := range records[1:] {
		row := make(CoralRow, len(headers))
		for i, h := range headers {
			if i < len(record) {
				row[h] = record[i]
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// RunQueryString runs a Coral query and returns the raw output string.
// Useful for embedding full context into the Claude summarization prompt.
func RunQueryString(sql string) (string, error) {
	cmd := exec.Command(".\\coral.exe", "sql", sql)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("coral sql failed: %w\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
