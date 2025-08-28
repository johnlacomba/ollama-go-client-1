# Ollama Go Client

This project is a Go client for communicating with a local Ollama LLM (Large Language Model). It provides a web interface with two modes: a standard "User-To-Model" chat and a "Model-To-Model" conversational interface.

## Project Structure

```
ollama-go-client-1
├── src
│   ├── static
│   │   └── index.html       # Web UI for both chat modes
│   ├── ollamaAPIWrapper
│   │   └── ollamaAPIWrapper.go # Wrapper for Ollama API calls like listing models
│   └── server
│       └── server.go        # HTTP server and API endpoint logic
├── go.mod
└── README.md
```

## Features

### General
- **Dual Interfaces**: Switch between a direct chat with a model and a conversation between two models.
- **Token Streaming**: Responses are streamed token-by-token using Server-Sent Events (SSE).
- **Dynamic Model Loading**: The list of available models is fetched directly from your local Ollama instance.
- **In-Memory Chat History**: The server maintains chat history for each user session (keyed by remote address).
- **Performance Metrics**: Each generated response displays the time it took to generate.
- **Conditional Auto-Scroll**: The chat window only auto-scrolls if you are already at the bottom, allowing you to scroll up and read previous messages during generation.

### User-To-Model Interface
- **Standard Chat**: A familiar chat interface for interacting with any available Ollama model.
- **Image Support**: Paste images directly into the prompt input to conduct multi-modal conversations.

### Model-To-Model Interface
- **Automated Conversations**: Pit two models against each other, starting with an initial prompt you provide.
- **Start/Stop Control**: Initiate and terminate the model-to-model conversation at any time.
- **Clean Sessions**: Each new conversation automatically clears the previous session history on the server.

## How Streaming Works

1. The browser sends a `POST` request to `/api/chat` with a JSON payload containing the model, prompt, and optional image.
2. The server forwards the full chat history for the session, plus the new prompt, to Ollama's `/api/chat` endpoint with `"stream": true`.
3. The server streams each partial chunk from Ollama back to the client as an SSE `data:` line, formatted as a JSON object (e.g., `{"token": "..."}`).
4. A final `event: done` message is sent, which includes the total generation duration.
5. The frontend client receives these events, accumulates the tokens into the latest message, and updates the UI in real-time.

## API Endpoints

- `GET /`  
  Serves the main `index.html` application.

- `POST /api/chat`  
  Accepts a JSON body and initiates an SSE stream for the model's response.
  - **Request Body**: `{"model": "<name>", "prompt": "<text>", "image": "<base64_string>"}`
  - **Response Stream**:
    - `data: {"token":"<chunk>"}`
    - `event: done\ndata: {"duration":"<Go duration string>"}`

- `POST /api/clear-history`  
  Clears the chat history for the current user session. No request body needed.

- `GET /api/tags`  
  Returns a JSON array of available model names from the local Ollama instance.

## Setup

```sh
# Ensure all dependencies are present
go mod tidy

# Run the server
go run src/server/server.go
```

Visit: `http://localhost:8080`

Ensure Ollama is running locally (default: `http://localhost:11434`).

## Chat History Behavior

- Stored per remote address. This is a simple approach for the demo; multiple users on the same NAT may share history.
- History is not persisted. Restarting the server clears all chat histories.
- The Model-To-Model interface automatically calls `/api/clear-history` to ensure each conversation is fresh.
- An initial two-message history is seeded for new sessions to prime the model, but these messages are not rendered in the UI.

## Customization

- **Generation Parameters**: Adjust `temperature`, `top_p`, etc., in `src/server/server.go` inside the `/api/chat` handler.
- **Initial Priming**: Modify the initial seed messages in `src/ollamaAPIWrapper/ollamaAPIWrapper.go` in the `NewClient()` function.

## Requirements

- Go 1.18+
- A local Ollama installation.
