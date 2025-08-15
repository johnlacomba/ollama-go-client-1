package main

import (
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

		// Use a simple session key (e.g., client IP) for demo purposes
		sessionKey := r.RemoteAddr

		// Get or initialize chat history
		history, ok := chatHistories[sessionKey]
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
		chatHistories[sessionKey] = history

		duration := time.Since(startTime)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"text":%q,"duration":"%v"}`, response.Message.Content, duration)
	})

	// This is correct: use the package-level function
	http.HandleFunc("/api/tags", client.GetModels)

	log.Printf("Starting server on %s", port)
	http.ListenAndServe(port, nil)
}
