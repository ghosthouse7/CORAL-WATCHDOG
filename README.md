<div align="center">

# 🏴‍☠️ CORAL WATCHDOG

**The Ultimate Cross-Source DevOps Incident Agent.**

[![Stack - Go](https://img.shields.io/badge/Stack-Go_1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Stack - Docker](https://img.shields.io/badge/Stack-Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![Powered by - Coral](https://img.shields.io/badge/Powered_by-Coral_SQL-FF6B6B?style=for-the-badge)](https://github.com/withcoral/coral)
[![AI - DeepSeek](https://img.shields.io/badge/AI-DeepSeek_|_GPT-4D4D4E?style=for-the-badge&logo=github&logoColor=white)]()
[![License - MIT](https://img.shields.io/badge/License-MIT-success?style=for-the-badge)](#license)

*Built for the [Pirates of the Coral-bean Hackathon](https://www.wemakedevs.org/hackathons/coral) (May 25–31, 2026).*

---
</div>

## ⚡ What It Does

When a Docker container goes unhealthy or exits unexpectedly, Coral Watchdog springs into action:

1. **Detects** the incident by polling the Docker Engine API every 60 seconds.
2. **Queries** Docker containers + GitHub pull requests in a single cross-source Coral SQL JOIN.
3. **Analyzes** the incident using DeepSeek AI (via GitHub Models — 100% free).
4. **Alerts** your team in Slack with a full, actionable root cause summary.
5. **Displays** everything in a live, animated dark-mode dashboard at `http://localhost:8080`.

> **The Magic:** Docker + GitHub + Slack. One single SQL query. Zero ETL pipelines required.

### 🌟 The Star Query

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


```python
content = """<div align="center">

# 🏴‍☠️ CORAL WATCHDOG

**The Ultimate Cross-Source DevOps Incident Agent.**

[![Stack - Go](https://img.shields.io/badge/Stack-Go_1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Stack - Docker](https://img.shields.io/badge/Stack-Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![Powered by - Coral](https://img.shields.io/badge/Powered_by-Coral_SQL-FF6B6B?style=for-the-badge)](https://github.com/withcoral/coral)
[![AI - DeepSeek](https://img.shields.io/badge/AI-DeepSeek_|_GPT-4D4D4E?style=for-the-badge&logo=github&logoColor=white)]()
[![License - MIT](https://img.shields.io/badge/License-MIT-success?style=for-the-badge)](#license)

*Built for the [Pirates of the Coral-bean Hackathon](https://www.wemakedevs.org/hackathons/coral) (May 25–31, 2026).*

---
</div>

## ⚡ What It Does

When a Docker container goes unhealthy or exits unexpectedly, Coral Watchdog springs into action:

1. **Detects** the incident by polling the Docker Engine API every 60 seconds.
2. **Queries** Docker containers + GitHub pull requests in a single cross-source Coral SQL JOIN.
3. **Analyzes** the incident using DeepSeek AI (via GitHub Models — 100% free).
4. **Alerts** your team in Slack with a full, actionable root cause summary.
5. **Displays** everything in a live, animated dark-mode dashboard at `http://localhost:8080`.

> **The Magic:** Docker + GitHub + Slack. One single SQL query. Zero ETL pipelines required.

### 🌟 The Star Query


```

```text
File generated.

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

---

## 🏗️ Architecture

```mermaid
graph TD;
    A[Docker Engine API] -->|Custom Source Spec| D(Coral SQL Layer);
    B[GitHub API] -->|Built-in Source| D;
    C[Slack API] -->|Built-in Source| D;
    
    D --> E{Go Agent};
    
    E -->|60s Polling| F[Watcher];
    E -->|Real-time UI| G[Web Dashboard];
    
    F -->|Context| H((gpt-4o-mini/Github Models));
    H -->|Root Cause Analysis| I[Slack #incidents Alert];

```

---

## 🎯 Bounties Targeted

| Bounty | Description | Status |
| --- | --- | --- |
| 🥇 **Captain's Bounty** | Best Enterprise Agent | 🚢 Submitted |
| 🐠 **Source Spec** | Custom Docker source spec | ✅ Built |
| ⌨️ **Captain's Log** | Blog post walkthrough | ✅ Written |
| 📦 **Social Swag** | LinkedIn/X post | ✅ Posted |

---

## 🛠️ Tech Stack

* **Go 1.22+** — High-performance agent, HTTP server, and Docker API client.
* **[Coral v0.3.0](https://github.com/withcoral/coral)** — The cross-source SQL engine powering the logic.
* **Docker Engine API** — Custom Coral source spec (written entirely from scratch).
* **GitHub & Slack APIs** — Leveraging Coral's built-in sources + webhooks.
* **DeepSeek-V3** — AI root cause analysis accessed via GitHub Models.
* **Vanilla Web** — HTML/CSS/JS dashboard featuring fluid animations and dark/light modes.

---

## 🚀 Quick Start

### Prerequisites

* Go 1.22+
* Docker Desktop (with TCP port `2375` enabled)
* Git

### 1. Clone & Setup

```bash
git clone [https://github.com/ghosthouse7/CORAL-WATCHDOG.git](https://github.com/ghosthouse7/CORAL-WATCHDOG.git)
cd CORAL-WATCHDOG

```

*Note: Download Coral v0.3.0 from the [official releases](https://github.com/withcoral/coral/releases) and place `coral.exe` (Windows) or `coral` (Linux/Mac) in the project root.*

### 2. Configure Environment

Create a `.env` file in the root directory:

```env
GITHUB_TOKEN=ghp_your_classic_token
GITHUB_OWNER=your-github-username
GITHUB_REPO=your-repo-name
SLACK_WEBHOOK_URL=[https://hooks.slack.com/services/xxx/yyy/zzz](https://hooks.slack.com/services/xxx/yyy/zzz)
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_INCIDENT_CHANNEL=incidents

```

### 3. Enable Docker TCP API

Navigate to **Docker Desktop → Settings → General** and enable:
`Expose daemon on tcp://localhost:2375`

### 4. Initialize Coral Sources (Windows Example)

```powershell
$env:GITHUB_TOKEN="ghp_..."
.\coral source add github

$env:SLACK_TOKEN="xoxb-..."
.\coral source add slack

.\coral source add --file .\sources\docker.yaml

```

### 5. Launch the Watchdog

```powershell
# Load env vars
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
    [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
  }
}

# Start the agent
go run main.go

```

🌐 **Open [http://localhost:8080](https://www.google.com/search?q=http://localhost:8080) in your browser to view the live dashboard.**

---

## 📊 Dashboard Features

* **📈 Overview** — Live metrics, container health, incident feed, and log stream.
* **🐳 Containers** — Full fleet view with CPU stats and visual status badges.
* **🤖 Incidents** — AI-generated root cause summaries with precise timestamps.
* **💬 Query Interface** — Natural language to Coral SQL translation with live results.

---

## 📂 Project Structure

```text
CORAL-WATCHDOG/
├── main.go              # Agent entrypoint + env validation
├── go.mod               # Go module dependencies
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
│   └── docker.yaml      # Custom Coral source spec for Docker 
├── queries/
│   └── incident.sql     # The star cross-source SQL query
└── .gitignore

```

---

### License

This project is licensed under the [MIT License](https://www.google.com/search?q=LICENSE).

*Made with ☕ and Coral SQL by [@ghosthouse7](https://github.com/ghosthouse7) during the Pirates of the Coral-bean Hackathon 2026.*
