package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings" // Import strings package
	"time"

	"ollama-go-client/src/ollamaAPIWrapper"
)

const port = ":8080"

var chatHistories = make(map[string][]ollamaAPIWrapper.Message)

// Define a struct for the incoming chat request payload
type chatRequestPayload struct {
	Model            string                   `json:"model"`
	Prompt           string                   `json:"prompt"`
	PersistentPrompt string                   `json:"persistent_prompt,omitempty"`
	Image            string                   `json:"image,omitempty"` // Base64 encoded image
	Options          ollamaAPIWrapper.Options `json:"options,omitempty"`
}

func main() {
	endpoint := "http://localhost:11434"
	timeout := 300 * time.Second

	// Set up a file server for the 'src/static' directory.
	// This will serve index.html, style.css, and any other static files.
	fs := http.FileServer(http.Dir("src/static"))
	http.Handle("/", fs)

	// API handlers
	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		// We must use POST to send a JSON body with an image
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		startTime := time.Now()

		// Decode the JSON payload from the request body
		var payload chatRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if payload.Prompt == "" && payload.Image == "" {
			http.Error(w, "A prompt or an image is required", http.StatusBadRequest)
			return
		}
		if payload.Model == "" {
			http.Error(w, "Model is required", http.StatusBadRequest)
			return
		}

		sessionKey := r.RemoteAddr
		history, ok := chatHistories[sessionKey]
		if !ok {
			history = ollamaAPIWrapper.NewClient(endpoint, timeout).ChatHistory
		}

		// Create the user message that will be saved in the long-term history.
		// This message is "clean" and does not contain the persistent prompt.
		historyMessage := ollamaAPIWrapper.Message{Role: "user", Content: payload.Prompt}
		if payload.Image != "" {
			base64Data := payload.Image
			if i := strings.Index(base64Data, ","); i != -1 {
				base64Data = base64Data[i+1:]
			}
			historyMessage.Images = []string{base64Data}
		}

		// For the current API call, create a temporary message that includes the persistent prompt.
		// This message is NOT saved in the long-term history.
		promptForOllama := payload.Prompt
		if payload.PersistentPrompt != "" {
			promptForOllama = payload.PersistentPrompt + "\n\n" + payload.Prompt
		}
		messageForOllama := ollamaAPIWrapper.Message{Role: "user", Content: promptForOllama, Images: historyMessage.Images}

		// The history sent to Ollama includes the long-term history plus the specially crafted message for this turn.
		messagesForOllama := append(history, messageForOllama)

		// Set headers for SSE streaming
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Prepare request body for Ollama
		reqBody := ollamaAPIWrapper.Request{
			Model:          payload.Model,
			StreamResponse: true,
			Messages:       messagesForOllama, // Use the temporary, combined message list
			Options:        payload.Options,   // Use options from the client
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

		// Stream tokens back to the client
		scanner := bufio.NewScanner(resp.Body)
		var assistantMsg string
		for scanner.Scan() {
			line := scanner.Bytes()
			var r ollamaAPIWrapper.Response
			if err := json.Unmarshal(line, &r); err != nil {
				continue
			}
			if r.Message.Content != "" {
				assistantMsg += r.Message.Content
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

		// Save the "clean" user message and the assistant's response to the long-term history.
		history = append(history, historyMessage)
		history = append(history, ollamaAPIWrapper.Message{Role: "assistant", Content: assistantMsg})
		chatHistories[sessionKey] = history

		// Send the final 'done' event
		duration := time.Since(startTime)
		donePayload := struct {
			Duration string `json:"duration"`
		}{Duration: duration.String()}
		b, _ := json.Marshal(donePayload)
		fmt.Fprintf(w, "event: done\ndata: %s\n\n", b)
		flusher.Flush()
	})

	http.HandleFunc("/api/clear-history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		sessionKey := r.RemoteAddr
		delete(chatHistories, sessionKey)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/tags", ollamaAPIWrapper.GetModels)
	log.Printf("Starting server on %s", port)
	http.ListenAndServe(port, nil)
}
