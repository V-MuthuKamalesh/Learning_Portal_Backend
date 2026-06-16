package judge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/collegeassess/backend/configs"
)

// ExecuteRequest is sent to the judge microservice.
type ExecuteRequest struct {
	Language      string `json:"language"`
	Source        string `json:"source"`
	Stdin         string `json:"stdin"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
}

// ExecuteResponse is returned by the judge microservice.
type ExecuteResponse struct {
	Status    string `json:"status"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	RuntimeMS int    `json:"runtime_ms"`
}

// Client calls the external judge service.
type Client struct {
	baseURL string
	enabled bool
	http    *http.Client
}

func NewClient(cfg configs.JudgeConfig) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.URL, "/"),
		enabled: cfg.Enabled,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Enabled() bool { return c.enabled }

func (c *Client) Execute(req ExecuteRequest) (*ExecuteResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("judge service is disabled")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Post(c.baseURL+"/execute", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("judge error: %s", string(raw))
	}
	var out ExecuteResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// NormalizeOutput trims whitespace for output comparison.
func NormalizeOutput(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}
	return strings.Join(lines, "\n")
}

// OutputsMatch compares judge output with expected output.
func OutputsMatch(actual, expected string) bool {
	return NormalizeOutput(actual) == NormalizeOutput(expected)
}
