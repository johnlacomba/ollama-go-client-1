package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ollama-go-client/src/client"
)

const port = ":8080"

var chatHistories = make(map[string][]client.Message)

func main() {
	// Read index.html content
	indexHTML, err := os.ReadFile("src/static/index.html")
	if err != nil {
		log.Fatal("Error reading index.html:", err)
	}
	indexHTMLStr := string(indexHTML)

	endpoint := "http://localhost:11434"
	timeout := 300 * time.Second

	// Set up HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, indexHTMLStr)
	})
	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		prompt := r.FormValue("prompt")
		model := r.FormValue("model")
		if prompt == "" || model == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Prompt and model are required")
			return
		}

		sessionKey := r.RemoteAddr
		history, ok := chatHistories[sessionKey]
		if !ok {
			history = client.NewClient(endpoint, timeout).ChatHistory
		}
		history = append(history, client.Message{Role: "user", Content: prompt})

		ollamaClient := client.NewClient(endpoint, timeout)
		ollamaClient.ChatHistory = history

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Prepare request body for streaming
		reqBody := client.Request{
			Model:          model,
			Prompt:         prompt,
			StreamResponse: true,
			Messages:       ollamaClient.ChatHistory,
			Options: client.Options{
				Temperature:      0.7,
				TopP:             0.95,
				FrequencyPenalty: 0.0,
				PresencePenalty:  0.0,
				MixtureSeed:      0,
				Seed:             0,
				BestOf:           0,
				Logprobs:         0,
			},
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
			return
		}

		httpClient := &http.Client{Timeout: timeout}
		resp, err := httpClient.Post(endpoint+"/api/chat", "application/json", bytes.NewBuffer(body))
		if err != nil {
			http.Error(w, "Failed to contact Ollama backend", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Stream tokens as they arrive
		scanner := bufio.NewScanner(resp.Body)
		var assistantMsg string
		for scanner.Scan() {
			line := scanner.Bytes()
			var r client.Response
			if err := json.Unmarshal(line, &r); err != nil {
				continue
			}
			if r.Message.Content != "" {
				assistantMsg += r.Message.Content

				// Wrap token in JSON so newlines / tabs are preserved safely
				chunk := struct {
					Token string `json:"token"`
				}{Token: r.Message.Content}

				b, _ := json.Marshal(chunk)
				fmt.Fprintf(w, "data: %s\n\n", b)
				flusher.Flush()
			}
			if r.Done {
				break
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Scanner error: %v", err)
		}

		// Save assistant's response to chat history
		history = append(history, client.Message{Role: "assistant", Content: assistantMsg})
		chatHistories[sessionKey] = history

		duration := time.Since(startTime)
		donePayload := struct {
			Duration string `json:"duration"`
		}{Duration: duration.String()}
		b, _ := json.Marshal(donePayload)
		fmt.Fprintf(w, "event: done\ndata: %s\n\n", b)
		flusher.Flush()
	})

	// This is correct: use the package-level function
	http.HandleFunc("/api/tags", client.GetModels)

	log.Printf("Starting server on %s", port)
	http.ListenAndServe(port, nil)
}
