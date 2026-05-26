package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// slackPayload is the Slack Incoming Webhook message body.
type slackPayload struct {
	Text        string        `json:"text"`
	Attachments []slackAttach `json:"attachments,omitempty"`
}

type slackAttach struct {
	Color  string       `json:"color"`
	Blocks []slackBlock `json:"blocks"`
}

type slackBlock struct {
	Type string     `json:"type"`
	Text *slackText `json:"text,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PostIncidentAlert sends a rich incident alert to the Slack #incidents channel.
func PostIncidentAlert(container Container, summary string) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL is not set")
	}

	name := ContainerName(container)
	ts := time.Now().Format("2006-01-02 15:04:05 UTC")

	payload := slackPayload{
		Text: fmt.Sprintf("🚨 *Incident Detected* — Container `%s` is `%s`", name, container.Status),
		Attachments: []slackAttach{
			{
				Color: "#FF0000",
				Blocks: []slackBlock{
					{
						Type: "section",
						Text: &slackText{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*🐳 Container:* `%s`\n*📦 Image:* `%s`\n*⚠️ Status:* `%s`\n*🕐 Detected:* %s",
								name, container.Image, container.Status, ts),
						},
					},
					{
						Type: "divider",
					},
					{
						Type: "section",
						Text: &slackText{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*🤖 AI Root Cause Analysis:*\n%s", summary),
						},
					},
					{
						Type: "divider",
					},
					{
						Type: "section",
						Text: &slackText{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*Container ID:* `%s`\n_Powered by Coral Watchdog 🏴‍☠️_", ShortID(container.ID)),
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack webhook post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// PostInfo sends a simple informational message to Slack.
func PostInfo(message string) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL is not set")
	}

	payload := slackPayload{Text: message}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
