// agent/query.go
package agent

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunIncidentQuery executes the Coral SQL query and returns the raw JSON result
func RunIncidentQuery() (string, error) {
	fmt.Println("🔍 Running cross-source Coral SQL query...")

	// This tells Go to run the Coral CLI in the background
	cmd := exec.Command("coral", "query", "--file", "queries/incident.sql", "--format", "json")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("❌ Coral query failed: %v\nDetails: %s", err, stderr.String())
	}

	fmt.Println("✅ Data fetched successfully from all 4 sources!")
	return out.String(), nil
}
