package main

import (
	"log"
	"os"
	"time"

	"github.com/yourusername/coral-watchdog/agent"
)

func main() {
	log.Println("🏴‍☠️  Coral Watchdog starting...")
	log.Println("   Watching Docker + GitHub + Slack for incidents")

	required := []string{
		"GITHUB_TOKEN",
		"GITHUB_OWNER",
		"GITHUB_REPO",
		"SLACK_WEBHOOK_URL",
		"SLACK_BOT_TOKEN",
		"SLACK_INCIDENT_CHANNEL",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			log.Fatalf("❌ Missing required environment variable: %s", env)
		}
	}

	w := agent.NewWatcher()

	log.Println("▶️  Running initial incident check...")
	if err := w.Check(); err != nil {
		log.Printf("⚠️  Check error: %v", err)
	}

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	log.Println("👁️  Watching for incidents every 60s... (Ctrl+C to stop)")
	for range ticker.C {
		if err := w.Check(); err != nil {
			log.Printf("⚠️  Check error: %v", err)
		}
	}
}
