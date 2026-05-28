package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const ghModelsURL = "https://models.inference.ai.azure.com/chat/completions"
const ghModel = "gpt-4o-mini"

type IncidentContext struct {
	ContainerName    string
	ContainerState   string
	Logs             []string
	CPUPercent       float64
	MemoryMB         float64
	CoralQueryResult string
}

func callGHModels(system, user string) (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", fmt.Errorf("GITHUB_TOKEN not set")
	}

	body := map[string]interface{}{
		"model":      ghModel,
		"max_tokens": 500,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", ghModelsURL, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	// Log status for debugging
	if resp.StatusCode != 200 {
		log.Printf("  ⚠️  GitHub Models HTTP %d: %s", resp.StatusCode, string(respBytes))
		return "", fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("JSON parse error: %w — body: %s", err, string(respBytes))
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response — body: %s", string(respBytes))
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func Summarize(ctx IncidentContext) (string, error) {
	system := `You are a DevOps incident analyst. Given a crashed container's logs and stats, 
write a concise root cause analysis in 3-5 sentences. Be specific and actionable.
Format: start with the most likely cause, then what to check next.`

	user := fmt.Sprintf(`Container: %s
State: %s
CPU: %.1f%%  Memory: %.1fMB

Last logs:
%s

Coral cross-source query result:
%s`,
		ctx.ContainerName,
		ctx.ContainerState,
		ctx.CPUPercent,
		ctx.MemoryMB,
		strings.Join(ctx.Logs, "\n"),
		ctx.CoralQueryResult,
	)

	return callGHModels(system, user)
}
