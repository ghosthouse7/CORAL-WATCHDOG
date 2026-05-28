package main

import (
	"log"
	"os"
	"time"

	"github.com/yourusername/coral-watchdog/agent"
)

func main() {
	log.Println("🏴‍☠️  Coral Watchdog starting...")

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

	// Start the dashboard web server in background
	go agent.StartDashboard(agent.NewDockerClient())

	log.Println("▶️  Running initial incident check...")
	if err := w.Check(); err != nil {
		log.Printf("⚠️  Check error: %v", err)
	}

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	log.Println("👁️  Watching every 60s · Dashboard at http://localhost:8080")
	for range ticker.C {
		if err := w.Check(); err != nil {
			log.Printf("⚠️  Check error: %v", err)
		}
	}
}
