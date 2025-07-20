package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	model      = "claude-3-haiku-20240307"
	apiKeyPath = "apikey.txt"
	logPath    = "chatlog.txt"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestPayload struct {
	Messages []Message `json:"messages"`
}

type ClaudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeResponse struct {
	Content []Content `json:"content"`
}

func main() {
	http.HandleFunc("/api/chat", chatHandler)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Claude API running"))
})

	fmt.Println("✅ Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		// Preflight CORS request
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	apiKey, err := os.ReadFile(apiKeyPath)
	if err != nil {
		http.Error(w, "API key missing", http.StatusInternalServerError)
		return
	}

	reqData := ClaudeRequest{
		Model:     model,
		MaxTokens: 1000,
		Messages:  payload.Messages,
	}

	jsonBody, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", strings.TrimSpace(string(apiKey)))
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Claude request failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		http.Error(w, string(bodyBytes), resp.StatusCode)
		return
	}

	var claudeResp ClaudeResponse
	json.Unmarshal(bodyBytes, &claudeResp)

	reply := ""
	for _, part := range claudeResp.Content {
		reply += part.Text
	}

	logToFile(payload.Messages, reply)

	json.NewEncoder(w).Encode(map[string]string{"reply": reply})
}

func logToFile(messages []Message, reply string) {
	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	writer := bufio.NewWriter(f)
	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("[%s] %s: %s\n", time.Now().Format(time.RFC822), msg.Role, msg.Content))
	}
	writer.WriteString(fmt.Sprintf("[%s] assistant: %s\n\n", time.Now().Format(time.RFC822), reply))
	writer.Flush()
}
