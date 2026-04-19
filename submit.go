package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// POST JSON to TYPA_API_URL/api/scores with Authorization Bearer TYPA_API_SECRET.
// TYPA_CLIENT_IP is optional (e.g. set by an SSH wrapper); defaults to 127.0.0.1 for local play.
func submitScore(username string, score int, wpm float64, ip string) {
	base := strings.TrimSuffix(os.Getenv("TYPA_API_URL"), "/")
	secret := os.Getenv("TYPA_API_SECRET")
	if base == "" || secret == "" {
		return
	}
	if ip == "" {
		ip = "127.0.0.1"
	}

	body, err := json.Marshal(map[string]any{
		"username": username,
		"score":    score,
		"wpm":      wpm,
		"ip":       ip,
	})
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, base+"/api/scores", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secret)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "typa: leaderboard submit: %v\n", err)
		return
	}
	_ = resp.Body.Close()
}
