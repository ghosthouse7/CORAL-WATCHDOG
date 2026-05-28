package agent

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Watcher struct {
	docker   *DockerClient
	seen     map[string]time.Time
	cooldown time.Duration
}

func NewWatcher() *Watcher {
	return &Watcher{
		docker:   NewDockerClient(),
		seen:     make(map[string]time.Time),
		cooldown: 10 * time.Minute,
	}
}

func (w *Watcher) Check() error {
	log.Println("🔍 Checking for unhealthy containers...")

	containers, err := w.docker.UnhealthyContainers()
	if err != nil {
		log.Printf("⚠️  Check error: list unhealthy containers: %v", err)
		return err
	}

	if len(containers) == 0 {
		log.Println("✅ All containers healthy")
		return nil
	}

	log.Printf("⚠️  Found %d unhealthy container(s)", len(containers))

	for _, c := range containers {
		name := ContainerName(c)

		if last, ok := w.seen[name]; ok && time.Since(last) < w.cooldown {
			log.Printf("  ⏳ Skipping %s (cooldown, last alerted %s ago)", name, time.Since(last).Round(time.Minute))
			continue
		}

		log.Printf("  🚨 Processing incident for: %s (%s)", name, c.Status)
		w.seen[name] = time.Now()

		// Fetch logs
		log.Printf("  📋 Fetching logs for %s...", name)
		logs, err := w.docker.GetLogs(c.ID, 50)
		if err != nil {
			logs = []string{"could not fetch logs: " + err.Error()}
		}

		// Fetch stats
		log.Printf("  📊 Fetching stats for %s...", name)
		stats, err := w.docker.GetStats(c.ID)
		if err != nil {
			stats = &ContainerStats{}
		}

		// Run Coral cross-source query with quoted column names
		log.Printf("  🪸  Running Coral cross-source query...")
		owner := os.Getenv("GITHUB_OWNER")
		repo := os.Getenv("GITHUB_REPO")
		crossSQL := fmt.Sprintf(
			`SELECT dc."Id", dc."Image", dc."Status", gp.title AS last_pr, gp.user__login AS author `+
				`FROM docker.containers dc `+
				`LEFT JOIN github.pulls gp ON gp.owner = '%s' AND gp.repo = '%s' `+
				`WHERE dc."State" != 'running' LIMIT 10`,
			owner, repo,
		)

		coralResult, err := RunQueryRaw(crossSQL)
		if err != nil {
			log.Printf("  ⚠️  Coral query error: %v", err)
			coralResult = "(coral query unavailable)"
		}

		// Build context for AI
		memMB := 0.0
		if stats.MemoryLimit > 0 {
			memMB = float64(stats.MemoryUsage) / 1024 / 1024
		}

		ctx := IncidentContext{
			ContainerName:    name,
			ContainerState:   c.State,
			Logs:             logs,
			CPUPercent:       stats.CPUPercent,
			MemoryMB:         memMB,
			CoralQueryResult: coralResult,
		}

		// Ask AI for root cause
		log.Printf("  🤖 Asking AI for root cause analysis...")
		summary, err := Summarize(ctx)
		if err != nil {
			summary = "AI analysis unavailable: " + err.Error()
		}

		// Add to dashboard
		AddIncident(c, summary)

		// Post to Slack
		log.Printf("  📣 Posting alert to Slack...")
		err = PostIncidentAlert(c, summary)
		if err != nil {
			log.Printf("  ❌ Slack error: %v", err)
		} else {
			log.Printf("  ✅ Incident alert posted for %s", name)
		}
	}

	return nil
}

func (w *Watcher) Run(interval time.Duration) {
	w.Check()
	log.Printf("👁️  Watching every %s · Dashboard at http://localhost:8080", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		w.Check()
	}
}

// logsPreview joins log lines into a single string, truncated.
func logsPreview(logs []string, max int) string {
	joined := strings.Join(logs, "\n")
	if len(joined) > max {
		return joined[:max] + "...(truncated)"
	}
	return joined
}
