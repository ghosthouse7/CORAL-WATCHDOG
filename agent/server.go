package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Incident struct {
	Container string `json:"container"`
	Status    string `json:"status"`
	Time      string `json:"time"`
	Summary   string `json:"summary"`
}

type DashContainer struct {
	Name      string  `json:"name"`
	Image     string  `json:"image"`
	Status    string  `json:"status"`
	Unhealthy bool    `json:"unhealthy"`
	CPU       float64 `json:"cpu,omitempty"`
}

var (
	incidentsMu sync.RWMutex
	incidents   []Incident
)

func AddIncident(container Container, summary string) {
	incidentsMu.Lock()
	defer incidentsMu.Unlock()
	inc := Incident{
		Container: ContainerName(container),
		Status:    container.Status,
		Time:      time.Now().UTC().Format("15:04 UTC"),
		Summary:   summary,
	}
	incidents = append([]Incident{inc}, incidents...)
	if len(incidents) > 10 {
		incidents = incidents[:10]
	}
}

func questionToSQL(question string) (string, error) {
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	system := fmt.Sprintf(`You are a Coral SQL expert. Convert natural language to Coral SQL.

AVAILABLE TABLES AND EXACT COLUMNS (never use SELECT *):
- docker.containers  → "Id", "Image", "Status", "State", "Created", "Command"
- docker.running     → "Id", "Image", "Status", "State", "Created"
- github.pulls       → title, state, merged_at, user__login  (MUST filter: owner='%s' AND repo='%s')
- slack.channels     → id, name

IMPORTANT: docker column names are case-sensitive and MUST be double-quoted.

EXAMPLE QUERIES (copy this exact style):
Q: show all containers
A: SELECT "Id", "Image", "Status", "State" FROM docker.containers LIMIT 20

Q: which containers are exited
A: SELECT "Id", "Image", "Status", "State" FROM docker.containers WHERE "State" = 'exited' LIMIT 20

Q: show running containers
A: SELECT "Id", "Image", "Status", "State" FROM docker.running LIMIT 20

Q: list recent pull requests
A: SELECT title, state, user__login, merged_at FROM github.pulls WHERE owner='%s' AND repo='%s' LIMIT 20

Q: show open PRs
A: SELECT title, state, user__login FROM github.pulls WHERE owner='%s' AND repo='%s' AND state='open' LIMIT 20

Q: list slack channels
A: SELECT id, name FROM slack.channels LIMIT 20

RULES:
- Output ONLY the raw SQL — no markdown, no backticks, no explanation, no semicolons
- NEVER use SELECT *
- Always double-quote docker column names: "Id", "Image", "Status", "State"
- For github tables ALWAYS include owner='%s' AND repo='%s' in WHERE`,
		owner, repo, owner, repo, owner, repo, owner, repo)

	sql, err := callGHModels(system, question)
	if err != nil {
		return "", err
	}

	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimRight(strings.TrimSpace(sql), ";")

	if strings.Contains(strings.ToUpper(sql), "SELECT *") {
		if strings.Contains(sql, "docker.containers") {
			sql = strings.Replace(sql, "SELECT *", `SELECT "Id", "Image", "Status", "State"`, 1)
		} else if strings.Contains(sql, "docker.running") {
			sql = strings.Replace(sql, "SELECT *", `SELECT "Id", "Image", "Status", "State"`, 1)
		} else if strings.Contains(sql, "github.pulls") {
			sql = strings.Replace(sql, "SELECT *", "SELECT title, state, user__login, merged_at", 1)
		} else if strings.Contains(sql, "slack.channels") {
			sql = strings.Replace(sql, "SELECT *", "SELECT id, name", 1)
		}
	}

	log.Printf("Generated SQL: %s", sql)
	return sql, nil
}

func jsonHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		fn(w, r)
	}
}

func StartDashboard(docker *DockerClient) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/containers", jsonHandler(func(w http.ResponseWriter, r *http.Request) {
		conts, err := docker.ListContainers()
		if err != nil {
			w.WriteHeader(503)
			json.NewEncoder(w).Encode(map[string]string{"error": "docker unavailable: " + err.Error()})
			return
		}
		var dash []DashContainer
		for _, c := range conts {
			name := ContainerName(c)
			s := strings.ToLower(c.Status)
			unhealthy := strings.Contains(s, "unhealthy") || strings.Contains(s, "exited") ||
				strings.Contains(s, "dead") || strings.Contains(s, "restarting")
			dc := DashContainer{Name: name, Image: c.Image, Status: c.Status, Unhealthy: unhealthy}
			if stats, err := docker.GetStats(c.ID); err == nil {
				dc.CPU = float64(int(stats.CPUPercent*10)) / 10
			}
			dash = append(dash, dc)
		}
		if dash == nil {
			dash = []DashContainer{}
		}
		json.NewEncoder(w).Encode(dash)
	}))

	mux.HandleFunc("/api/incidents", jsonHandler(func(w http.ResponseWriter, r *http.Request) {
		incidentsMu.RLock()
		defer incidentsMu.RUnlock()
		out := incidents
		if out == nil {
			out = []Incident{}
		}
		json.NewEncoder(w).Encode(out)
	}))

	mux.HandleFunc("/api/query", jsonHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		var body struct {
			Question string `json:"question"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Question == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "question is required"})
			return
		}

		log.Printf("NL Query: %q", body.Question)

		sql, err := questionToSQL(body.Question)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		results, err := RunQueryRows(sql)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"sql": sql, "error": err.Error()})
			return
		}
		if results == nil {
			results = []map[string]string{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sql":     sql,
			"results": results,
		})
	}))

	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	log.Println("🌐 Dashboard → http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Dashboard error: %v", err)
	}
}
