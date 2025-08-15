package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ollama-go-client/src/client"
)

const port = ":8080"

var chatHistories = make(map[string]map[string][]client.Message)

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

		// Get or initialize per-model chat history for this session
		modelHistories, ok := chatHistories[sessionKey]
		if !ok {
			modelHistories = make(map[string][]client.Message)
		}
		history, ok := modelHistories[model]
		if !ok {
			history = client.NewClient(endpoint, timeout).ChatHistory
		}

		// Append the new user message
		history = append(history, client.Message{Role: "user", Content: prompt})

		// Create a client with this history
		ollamaClient := client.NewClient(endpoint, timeout)
		ollamaClient.ChatHistory = history

		response, err := ollamaClient.SendRequest(
			model, prompt,
			0.7, 0.95, 0.0, 0.0, 0, 0, 0, 0,
		)
		if err != nil {
			log.Printf("Error generating response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error generating response")
			return
		}

		// Append assistant's response to history
		history = append(history, client.Message{Role: "assistant", Content: response.Message.Content})
		modelHistories[model] = history
		chatHistories[sessionKey] = modelHistories

		duration := time.Since(startTime)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"text":%q,"duration":"%v","history":%s}`, response.Message.Content, duration, toJSON(history))
	})

	// This is correct: use the package-level function
	http.HandleFunc("/api/tags", client.GetModels)

	log.Printf("Starting server on %s", port)
	http.ListenAndServe(port, nil)
}

// Helper to marshal history to JSON
func toJSON(history []client.Message) string {
	b, _ := json.Marshal(history)
	return string(b)
}
