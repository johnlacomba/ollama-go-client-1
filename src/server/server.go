package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings" // Restoring the strings import
	"time"

	"ollama-go-client/src/ollamaAPIWrapper"
)

const port = ":8080"

// Restore server-side history management
var chatHistories = make(map[string][]ollamaAPIWrapper.Message)

// Restore the original payload structure
type chatRequestPayload struct {
	Model            string                   `json:"model"`
	Prompt           string                   `json:"prompt"`
	PersistentPrompt string                   `json:"persistent_prompt,omitempty"`
	Image            string                   `json:"image,omitempty"`
	Options          ollamaAPIWrapper.Options `json:"options,omitempty"`
	SummarizeHistory bool                     `json:"summarize_history,omitempty"` // New field for M2M chat
}

// getSummary makes a blocking, non-streaming call to Ollama to summarize a chat history.
func getSummary(endpoint string, timeout time.Duration, modelName string, history []ollamaAPIWrapper.Message) (string, error) {
	const maxHistoryForSummary = 10
	if len(history) > maxHistoryForSummary {
		history = history[len(history)-maxHistoryForSummary:]
	}

	var historyText strings.Builder
	for _, msg := range history {
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		historyText.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	summarizerPrompt := fmt.Sprintf("Concisely summarize the key points of the following conversation exchange:\n\n%s", historyText.String())

	reqBody := ollamaAPIWrapper.Request{
		Model:          modelName, // Use the same model for summarization
		StreamResponse: false,     // We need the full summary at once
		Messages:       []ollamaAPIWrapper.Message{{Role: "user", Content: summarizerPrompt}},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{Timeout: timeout}
	resp, err := httpClient.Post(endpoint+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var summaryResponse ollamaAPIWrapper.Response
	if err := json.NewDecoder(resp.Body).Decode(&summaryResponse); err != nil {
		return "", err
	}

	return summaryResponse.Message.Content, nil
}

func main() {
	endpoint := "http://localhost:11434"
	timeout := 300 * time.Second

	fs := http.FileServer(http.Dir("src/static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload chatRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		sessionKey := r.RemoteAddr // Simple session management
		history, ok := chatHistories[sessionKey]
		if !ok {
			history = []ollamaAPIWrapper.Message{}
		}

		// The message to be saved in history (clean, without any special prompting)
		historyMessage := ollamaAPIWrapper.Message{Role: "user", Content: payload.Prompt}
		if payload.Image != "" {
			base64Data := payload.Image
			if i := strings.Index(base64Data, ","); i != -1 {
				base64Data = base64Data[i+1:]
			}
			historyMessage.Images = []string{base64Data}
		}

		var messagesForOllama []ollamaAPIWrapper.Message
		promptForOllama := payload.Prompt

		// --- New Summarization and Prompting Logic ---
		if payload.SummarizeHistory && len(history) > 0 {
			// 1. Get a summary of the recent history.
			summaryContext, err := getSummary(endpoint, timeout, payload.Model, history)
			if err != nil {
				log.Printf("Could not generate summary, proceeding without it: %v", err)
				// Fallback to using the full history if summarization fails
				messagesForOllama = append(history, ollamaAPIWrapper.Message{Role: "user", Content: promptForOllama})
			} else {
				// 2. Build a new structured prompt with the summary.
				structuredPrompt := fmt.Sprintf(
					"Here is a summary of the recent conversation:\n\"%s\"\n\nBased on that summary, continue the conversation by responding to the following prompt: \"%s\"",
					summaryContext,
					payload.Prompt,
				)
				promptForOllama = structuredPrompt
				// The history is now the summary, so we send only the new structured prompt.
				messagesForOllama = []ollamaAPIWrapper.Message{{Role: "user", Content: promptForOllama}}
			}
		} else {
			// Default behavior for User-to-Model chat: use the full history.
			messagesForOllama = append(history, ollamaAPIWrapper.Message{Role: "user", Content: promptForOllama})
		}

		// 3. Emphatically apply the persistent prompt to the final prompt string.
		if payload.PersistentPrompt != "" {
			finalPromptContent := fmt.Sprintf(
				"Your primary instruction is: %s\n\n---\n\n%s",
				payload.PersistentPrompt,
				promptForOllama,
			)
			// Replace the last message's content with this new, fully-formed prompt.
			if len(messagesForOllama) > 0 {
				messagesForOllama[len(messagesForOllama)-1].Content = finalPromptContent
			} else {
				// Handle case where there's no history (e.g., first message with summary)
				messagesForOllama = append(messagesForOllama, ollamaAPIWrapper.Message{Role: "user", Content: finalPromptContent})
			}
		}
		// --- End of New Logic ---

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		reqBody := ollamaAPIWrapper.Request{
			Model:          payload.Model,
			StreamResponse: true,
			Messages:       messagesForOllama,
			Options:        payload.Options,
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

		scanner := bufio.NewScanner(resp.Body)
		var assistantMsg string
		var finalOllamaResponse ollamaAPIWrapper.Response
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
				finalOllamaResponse = r
				break
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Scanner error: %v", err)
		}

		// Append the clean user prompt and the full assistant response to history
		history = append(history, historyMessage)
		history = append(history, ollamaAPIWrapper.Message{Role: "assistant", Content: assistantMsg})
		chatHistories[sessionKey] = history

		// Send the final 'done' event using the accurate duration from Ollama
		duration := time.Duration(finalOllamaResponse.TotalDuration)
		donePayload := struct {
			Duration string `json:"duration"`
		}{Duration: duration.String()}
		b, _ := json.Marshal(donePayload)
		fmt.Fprintf(w, "event: done\ndata: %s\n\n", b)
		flusher.Flush()
	})

	// Restore the functional /api/clear-history endpoint
	http.HandleFunc("/api/clear-history", func(w http.ResponseWriter, r *http.Request) {
		sessionKey := r.RemoteAddr
		delete(chatHistories, sessionKey)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/tags", ollamaAPIWrapper.GetModels)
	log.Printf("Starting server on %s", port)
	// This is the fix. log.Fatal will print any error that occurs on startup
	// and will correctly block the main function from exiting.
	log.Fatal(http.ListenAndServe(port, nil))
}
