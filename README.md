# Ollama Go Client

This project is a Go client for communicating with a local Ollama LLM (Large Language Model). It provides a web interface to send prompts and receive *streaming* responses from the model.

## Project Structure

```
ollama-go-client-1
├── src
│   ├── static
│   │   └── index.html       # Web UI (SSE streaming)
│   ├── client
│   │   └── client.go        # Client wrapper (chat history + request builder)
│   └── server
│       └── server.go        # HTTP server + /api/chat (SSE) + /api/tags
├── go.mod
└── README.md
```

## Features

- Web-based chat interface.
- Token-by-token streaming via Server-Sent Events (SSE) from `/api/chat`.
- Per-session in‑memory chat history (clears on page reload or server restart).
- Full chat history (including the first two seeded messages) always sent as context.
- First two seeded messages (system-style priming) are *not* rendered to the user.
- Model list dynamically loaded from `/api/tags`.
- Auto-expanding prompt input; Enter submits (Shift+Enter for newline blocked).
- Lightweight, dependency-free (std lib + Ollama).

## How Streaming Works

1. Browser opens an `EventSource` to:  
   `GET /api/chat?model=<model>&prompt=<urlencoded prompt>`
2. Server forwards full chat history + new prompt to Ollama's `/api/chat` with `"stream": true`.
3. Server streams each partial chunk back as `data:` lines.
4. A final `event: done` message includes total generation duration.
5. Frontend accumulates chunks into the latest assistant message.

## API Endpoints

- `GET /`  
  Serves `index.html`.
- `GET /api/tags`  
  Returns JSON array of available model names.
- `GET /api/chat?model=<name>&prompt=<text>`  
  SSE stream of model output (token chunks). Final event named `done` contains:  
  `{"duration":"<Go duration string>"}`

Example SSE stream (simplified):
```
data: Hel
data: lo 
data: world
event: done
data: {"duration":"1.237s"}
```

## Setup

```sh
go mod tidy
go run src/server/server.go
```

Visit: `http://localhost:8080`

Ensure Ollama is running locally (default: `http://localhost:11434`).

## Chat History Behavior

- Stored per remote address only (simple demo approach).
- Not persisted; restarting or reloading resets context.
- Hidden seed messages still influence generation.

## Customization

Adjust generation parameters in `server.go` (inside `/api/chat` handler):
```
Temperature: 0.7
TopP:        0.95
...
```

Modify initial priming messages in `client.NewClient()`.

## Notes / Future Ideas

- Add real session IDs (e.g., cookie or UUID) instead of `RemoteAddr`.
- Add per-model independent histories (map[session][model]) if needed.
- Add cancel support (client closes EventSource).
- Parameter controls in UI (temperature, top_p, etc.).
- Persist history (Redis / BoltDB) if required.

## Requirements

- Go 1.18+
- Local Ollama installation and
