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

	unhealthy, err := w.docker.UnhealthyContainers()
	if err != nil {
		return fmt.Errorf("list unhealthy containers: %w", err)
	}

	if len(unhealthy) == 0 {
		log.Println("✅ All containers healthy")
		return nil
	}

	log.Printf("⚠️  Found %d unhealthy container(s)", len(unhealthy))

	for _, c := range unhealthy {
		name := ContainerName(c)

		if last, ok := w.seen[c.ID]; ok {
			if time.Since(last) < w.cooldown {
				log.Printf("  ⏳ Skipping %s (cooldown, last alerted %s ago)", name, time.Since(last).Round(time.Second))
				continue
			}
		}

		log.Printf("  🚨 Processing incident for: %s (%s)", name, c.Status)
		w.seen[c.ID] = time.Now()

		if err := w.handleIncident(c); err != nil {
			log.Printf("  ❌ Failed to handle incident for %s: %v", name, err)
		}
	}

	return nil
}

func (w *Watcher) handleIncident(c Container) error {
	name := ContainerName(c)

	// 1. Get container logs
	log.Printf("  📋 Fetching logs for %s...", name)
	logs, err := w.docker.GetLogs(c.ID, 50)
	if err != nil {
		log.Printf("  ⚠️  Could not get logs: %v", err)
		logs = []string{"(logs unavailable)"}
	}

	// 2. Get container stats
	log.Printf("  📊 Fetching stats for %s...", name)
	var cpuPct, memMB float64
	stats, err := w.docker.GetStats(c.ID)
	if err != nil {
		log.Printf("  ⚠️  Could not get stats: %v", err)
	} else {
		cpuPct = stats.CPUPercent
		memMB = float64(stats.MemoryUsage) / 1024 / 1024
	}

	// 3. Run the Coral cross-source query
	log.Println("  🪸  Running Coral cross-source query...")
	coralResult, err := runIncidentQuery(name)
	if err != nil {
		log.Printf("  ⚠️  Coral query error: %v", err)
		coralResult = "(coral query unavailable — check source setup)"
	}

	// 4. Summarize with DeepSeek via GitHub Models
	log.Println("  🤖 Asking DeepSeek for root cause analysis...")
	ctx := IncidentContext{
		ContainerName:    name,
		ContainerState:   c.Status,
		Logs:             logs,
		CPUPercent:       cpuPct,
		MemoryMB:         memMB,
		CoralQueryResult: coralResult,
	}

	summary, err := Summarize(ctx)
	if err != nil {
		log.Printf("  ⚠️  Summarize error: %v", err)
		summary = fmt.Sprintf("⚠️ AI analysis unavailable: %v\n\nRaw logs:\n%s", err, strings.Join(logs, "\n"))
	}

	// 5. Post to Slack
	log.Println("  📣 Posting alert to Slack...")
	if err := PostIncidentAlert(c, summary); err != nil {
		return fmt.Errorf("post slack alert: %w", err)
	}

	log.Printf("  ✅ Incident alert posted for %s", name)
	return nil
}

func runIncidentQuery(containerName string) (string, error) {
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	// Join Docker containers with GitHub PRs
	// Slack messages table not available in free tier — using channels instead
	sql := fmt.Sprintf(`
SELECT
  dc.id                   AS container_id,
  dc.status               AS container_status,
  dc.image                AS image,
  gp.title                AS last_merged_pr,
  gp.merged_at            AS deploy_time,
  gp.author__login        AS deployed_by,
  sc.id                   AS slack_channel_id,
  sc.name                 AS slack_channel_name
FROM docker.containers dc
LEFT JOIN github.pulls gp
  ON gp.owner = '%s'
  AND gp.repo = '%s'
  AND gp.state = 'closed'
LEFT JOIN slack.channels sc
  ON sc.name = 'social'
WHERE dc.status LIKE '%%Exited%%'
   OR dc.status LIKE '%%unhealthy%%'
   OR dc.status LIKE '%%dead%%'
ORDER BY gp.merged_at DESC
LIMIT 20
`, owner, repo)

	return RunQueryString(sql)
}
