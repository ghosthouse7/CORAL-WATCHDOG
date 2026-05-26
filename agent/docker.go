package agent

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DockerClient calls the Docker Engine HTTP API.
// DOCKER_HOST defaults to http://localhost:2375 (socat proxy in docker-compose).
type DockerClient struct {
	BaseURL string
	http    *http.Client
}

// Container mirrors the relevant fields from /containers/json
type Container struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	Status  string            `json:"Status"`
	State   string            `json:"State"`
	Created int64             `json:"Created"`
	Labels  map[string]string `json:"Labels"`
}

// ContainerStats holds trimmed stats from /containers/{id}/stats?stream=false
type ContainerStats struct {
	Name        string
	CPUPercent  float64
	MemoryUsage uint64
	MemoryLimit uint64
}

func NewDockerClient() *DockerClient {
	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		host = "http://localhost:2375"
	}
	return &DockerClient{
		BaseURL: strings.TrimRight(host, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *DockerClient) get(path string) ([]byte, error) {
	resp, err := d.http.Get(d.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("docker api GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("docker api %s returned %d: %s", path, resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// ListContainers returns all containers (running + stopped).
func (d *DockerClient) ListContainers() ([]Container, error) {
	data, err := d.get("/containers/json?all=true")
	if err != nil {
		return nil, err
	}

	var containers []Container
	if err := json.Unmarshal(data, &containers); err != nil {
		return nil, fmt.Errorf("unmarshal containers: %w", err)
	}
	return containers, nil
}

// UnhealthyContainers returns containers whose status indicates a problem.
func (d *DockerClient) UnhealthyContainers() ([]Container, error) {
	all, err := d.ListContainers()
	if err != nil {
		return nil, err
	}

	var unhealthy []Container
	for _, c := range all {
		s := strings.ToLower(c.Status)
		if strings.Contains(s, "unhealthy") ||
			strings.Contains(s, "exited") ||
			strings.Contains(s, "dead") ||
			strings.Contains(s, "restarting") {
			unhealthy = append(unhealthy, c)
		}
	}
	return unhealthy, nil
}

// GetLogs returns the last N lines of logs from a container.
// Docker multiplexes stdout/stderr — we strip the 8-byte header.
func (d *DockerClient) GetLogs(containerID string, tail int) ([]string, error) {
	path := fmt.Sprintf(
		"/containers/%s/logs?stdout=1&stderr=1&timestamps=1&tail=%d",
		containerID, tail,
	)

	data, err := d.get(path)
	if err != nil {
		return nil, err
	}

	return parseMuxedLogs(data), nil
}

// parseMuxedLogs strips Docker's 8-byte stream header from log output.
// Format: [stream_type(1)] [0 0 0(3)] [size(4)] [payload...]
func parseMuxedLogs(data []byte) []string {
	var lines []string
	for len(data) >= 8 {
		size := binary.BigEndian.Uint32(data[4:8])
		if int(size) > len(data)-8 {
			break
		}
		line := strings.TrimRight(string(data[8:8+size]), "\n\r")
		if line != "" {
			lines = append(lines, line)
		}
		data = data[8+size:]
	}

	// Fallback: if parsing produced nothing, treat as plain text
	if len(lines) == 0 {
		for _, l := range strings.Split(string(data), "\n") {
			l = strings.TrimSpace(l)
			if l != "" {
				lines = append(lines, l)
			}
		}
	}

	return lines
}

// GetStats returns CPU and memory stats for a container.
func (d *DockerClient) GetStats(containerID string) (*ContainerStats, error) {
	data, err := d.get(fmt.Sprintf("/containers/%s/stats?stream=false", containerID))
	if err != nil {
		return nil, err
	}

	var raw struct {
		Name  string `json:"name"`
		CPUSt struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs     int    `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUSt struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemSt struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}

	// Calculate CPU %
	cpuDelta := float64(raw.CPUSt.CPUUsage.TotalUsage - raw.PreCPUSt.CPUUsage.TotalUsage)
	sysDelta := float64(raw.CPUSt.SystemCPUUsage - raw.PreCPUSt.SystemCPUUsage)
	cpuPercent := 0.0
	if sysDelta > 0 && cpuDelta > 0 {
		cpuPercent = (cpuDelta / sysDelta) * float64(raw.CPUSt.OnlineCPUs) * 100.0
	}

	return &ContainerStats{
		Name:        raw.Name,
		CPUPercent:  cpuPercent,
		MemoryUsage: raw.MemSt.Usage,
		MemoryLimit: raw.MemSt.Limit,
	}, nil
}

// ShortID returns the first 12 chars of a container ID.
func ShortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// ContainerName returns the primary name (strips leading slash).
func ContainerName(c Container) string {
	if len(c.Names) > 0 {
		return strings.TrimPrefix(c.Names[0], "/")
	}
	return ShortID(c.ID)
}
