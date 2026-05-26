package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const githubModelsAPI = "https://models.inference.ai.azure.com/chat/completions"
const model = "DeepSeek-V3-0324"

// IncidentContext is the full data bundle passed to the AI for summarization.
type IncidentContext struct {
	ContainerName    string
	ContainerState   string
	Logs             []string
	CPUPercent       float64
	MemoryMB         float64
	CoralQueryResult string
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Summarize calls GitHub Models (DeepSeek) to generate an incident summary.
func Summarize(ctx IncidentContext) (string, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return "", fmt.Errorf("GITHUB_TOKEN is not set")
	}

	logsSection := "No logs captured."
	if len(ctx.Logs) > 0 {
		logsSection = ""
		for _, l := range ctx.Logs {
			logsSection += "  " + l + "\n"
		}
	}

	userPrompt := fmt.Sprintf(`
An incident has been detected. Here is all the data:

## Affected Container
- Name: %s
- State: %s
- CPU: %.1f%%
- Memory: %.0f MB

## Recent Logs (last 50 lines)
%s

## Cross-Source Coral Query Result
(Joining Docker + GitHub + Slack)
%s

Please analyse this incident and provide:
1. **Root Cause** — What likely went wrong and why
2. **Impact** — What services or users are affected
3. **Timeline** — Key events leading up to this
4. **Recommended Fix** — Concrete steps the on-call engineer should take right now
5. **Prevention** — What to change to stop this happening again

Be concise, specific, and actionable. Use bullet points.
`,
		ctx.ContainerName,
		ctx.ContainerState,
		ctx.CPUPercent,
		ctx.MemoryMB,
		logsSection,
		ctx.CoralQueryResult,
	)

	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "You are a senior DevOps engineer analysing production incidents. Be concise, specific, and focus on actionable remediation. Only conclude from evidence provided.",
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodPost, githubModelsAPI, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+githubToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("api call failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBytes))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	return chatResp.Choices[0].Message.Content, nil
}
