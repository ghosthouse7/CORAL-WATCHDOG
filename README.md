# 🏴‍☠️ Coral Watchdog

> A DevOps incident agent that joins Docker + GitHub + Slack in a single Coral SQL query — and posts AI-powered root cause analysis to Slack when containers go down.

Built for the [Pirates of the Coral-bean Hackathon](https://www.wemakedevs.org/hackathons/coral) (May 25–31, 2026).

![Dashboard Preview](https://img.shields.io/badge/Stack-Go%20%7C%20Coral%20%7C%20Docker%20%7C%20DeepSeek-cyan?style=flat-square)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

---

## What It Does

When a Docker container goes unhealthy or exits unexpectedly:

1. **Detects** the incident by polling the Docker Engine API every 60 seconds
2. **Queries** Docker containers + GitHub pull requests in a single cross-source Coral SQL JOIN
3. **Analyzes** the incident using DeepSeek AI (via GitHub Models — free)
4. **Alerts** your team in Slack with a full root cause summary
5. **Displays** everything in a live animated dashboard at `http://localhost:8080`

### The Star Query

```sql
SELECT dc.image, dc.status, dc.state,
       gp.title  AS last_pr,
       gp.user__login AS author,
       sc.name   AS slack_channel
FROM   docker.containers dc
LEFT JOIN github.pulls gp
  ON  gp.owner = 'your-org' AND gp.repo = 'your-repo'
LEFT JOIN slack.channels sc ON 1=1
WHERE  dc.state != 'running'
LIMIT  20;
```

Docker + GitHub + Slack. One query. No ETL.

---

## Architecture

```
[Docker Engine API]  [GitHub API]  [Slack API]
        │                  │              │
        └──────────────────┴──────────────┘
                           │
                    Coral SQL Layer
                           │
                    Go Agent (main.go)
                    ┌──────┴──────────┐
               Dashboard           Watcher
            (localhost:8080)    (60s polling)
                                     │
                              gpt-4o-mini (GitHub Models)
                                     │
                              Slack #incidents alert
```

---

## Bounties Targeted

| Bounty | Description | Status |
|--------|-------------|--------|
| 🥇 Captain's Bounty (MacBook) | Best Enterprise Agent | Submitted |
| 🐠 Source Spec ($100) | Custom Docker source spec | ✅ Built |
| ⌨️ Captain's Log (Keychron) | Blog post walkthrough | ✅ Written |
| 📦 Social Swag | LinkedIn/X post | ✅ Posted |

---

## Tech Stack

- **Go 1.22+** — Agent, HTTP server, Docker API client
- **[Coral v0.3.0](https://github.com/withcoral/coral)** — Cross-source SQL engine
- **Docker Engine API** — Custom Coral source spec (written from scratch)
- **GitHub API** — Built-in Coral source
- **Slack API** — Built-in Coral source + webhook alerts
- **DeepSeek-V3** via GitHub Models — AI root cause analysis (free)
- **Vanilla HTML/CSS/JS** — Dashboard with dark/light mode

---

## Quick Start

### Prerequisites

- Go 1.22+
- Docker Desktop (with TCP port 2375 enabled)
- Git

### 1. Clone & Setup

```bash
git clone https://github.com/ghosthouse7/CORAL-WATCHDOG.git
cd CORAL-WATCHDOG
```

Download Coral v0.3.0 from https://github.com/withcoral/coral/releases
and place `coral.exe` (Windows) or `coral` (Linux/Mac) in the project root.

### 2. Create .env

```env
GITHUB_TOKEN=ghp_your_classic_token
GITHUB_OWNER=your-github-username
GITHUB_REPO=your-repo-name
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx/yyy/zzz
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_INCIDENT_CHANNEL=incidents
```

### 3. Enable Docker TCP API

Docker Desktop → Settings → General → Enable "Expose daemon on tcp://localhost:2375"

### 4. Add Coral Sources

```bash
# Windows
$env:GITHUB_TOKEN="ghp_..."
.\coral source add github

$env:SLACK_TOKEN="xoxb-..."
.\coral source add slack

.\coral source add --file .\sources\docker.yaml
```

### 5. Run

```bash
# Windows — load env vars first
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
    [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
  }
}

go run main.go
```

Open http://localhost:8080 in your browser.

---

## Dashboard Features

- **Overview** — Live metrics, container health, incident feed, log stream
- **Containers** — Full fleet view with CPU stats and status badges
- **Incidents** — AI-generated root cause summaries with timestamps
- **Query** — Natural language → Coral SQL → live results

---

## Project Structure

```
CORAL-WATCHDOG/
├── main.go              # Agent entrypoint + env validation
├── go.mod               # Go module
├── frontend/
│   └── index.html       # Dashboard (single file, no build step)
├── agent/
│   ├── coral.go         # Coral query runner
│   ├── docker.go        # Docker Engine API client
│   ├── server.go        # HTTP server + NL query endpoint
│   ├── slack.go         # Slack webhook notifier
│   ├── summarize.go     # DeepSeek AI integration
│   └── watcher.go       # Container health watcher
├── sources/
│   └── docker.yaml      # Custom Coral source spec for Docker ← bounty!
├── queries/
│   └── incident.sql     # The star cross-source SQL query
└── .gitignore
```

---

## License

MIT — see [LICENSE](LICENSE)

---

*Made with ☕ and Coral SQL by [@ghosthouse7](https://github.com/ghosthouse7) during the Pirates of the Coral-bean Hackathon 2026*
