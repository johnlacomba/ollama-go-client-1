package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	Endpoint    string
	Timeout     time.Duration
	ChatHistory []Message
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model            string    `json:"model"`
	Prompt           string    `json:"prompt"`
	StreamResponse   bool      `json:"stream,omitempty"`
	Messages         []Message `json:"messages,omitempty"`
	Temperature      float32   `json:"temperature,omitempty"` // Add temperature parameter
	TopP             float32   `json:"top_p,omitempty"`       // Add top_p parameter
	FrequencyPenalty float32   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32   `json:"presence_penalty,omitempty"`
	MixtureSeed      int64     `json:"mixture_seed,omitempty"`
	Seed             int64     `json:"seed,omitempty"`
	BestOf           int       `json:"best_of,omitempty"`
	Logprobs         int       `json:"logprobs,omitempty"`
	// Add other parameters as needed
}

type Response struct {
	Model              string  `json:"model"`
	CreatedAt          string  `json:"created_at"`
	Message            Message `json:"message"`
	Done               bool    `json:"done"`
	TotalDuration      int64   `json:"total_duration"`
	LoadDuration       int64   `json:"load_duration"`
	PromptEvalCount    int     `json:"prompt_eval_count"`
	PromptEvalDuration int64   `json:"prompt_eval_duration"`
	EvalCount          int     `json:"eval_count"`
	EvalDuration       int64   `json:"eval_duration"`
}

func NewClient(endpoint string, timeout time.Duration) *Client {
	return &Client{
		Endpoint: endpoint,
		Timeout:  timeout,
		ChatHistory: []Message{
			{Role: "user", Content: "INITIAL PROMPT"},
			{Role: "assistant", Content: "RESPONSE FROM INITIAL PROMPT"},
		},
	}
}

func (c *Client) SendRequest(model string, prompt string, temperature float32, topP float32, frequencyPenalty float32, presencePenalty float32, mixtureSeed int64, seed int64, bestOf int, logprobs int) (*Response, error) {
	// Add user's message to chat history
	c.ChatHistory = append(c.ChatHistory, Message{Role: "user", Content: prompt})

	reqBody := Request{
		Model:            model,
		Prompt:           prompt,
		StreamResponse:   true, // Enable streaming
		Temperature:      temperature,
		TopP:             topP,
		FrequencyPenalty: frequencyPenalty,
		PresencePenalty:  presencePenalty,
		MixtureSeed:      mixtureSeed,
		Seed:             seed,
		BestOf:           bestOf,
		Logprobs:         logprobs,
		Messages:         c.ChatHistory,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Post(c.Endpoint+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read streaming response line by line
	var finalResponse Response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		var r Response
		if err := json.Unmarshal(line, &r); err != nil {
			return nil, err
		}
		// Accumulate the content
		if r.Message.Content != "" {
			finalResponse.Message.Content += r.Message.Content
		}
		finalResponse.Done = r.Done
		if r.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Add assistant's response to chat history
	c.ChatHistory = append(c.ChatHistory, Message{Role: "assistant", Content: finalResponse.Message.Content})

	return &finalResponse, nil
}

func (c *Client) ListModels() ([]string, error) {
	resp, err := http.Get(c.Endpoint + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var modelsResponse struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	var modelNames []string
	for _, model := range modelsResponse.Models {
		modelNames = append(modelNames, model.Name)
	}

	return modelNames, nil
}
