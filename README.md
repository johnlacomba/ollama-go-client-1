# Ollama Go Client

This project is a Go client for communicating with a local Ollama LLM (Large Language Model). It provides a simple web interface to send prompts and receive responses from the model.

## Project Structure

```
ollama-go-client-1
├── src
│   ├── static
│   │   └── index.html       # Web UI for interacting with the LLM
│   ├── client
│   │   └── client.go        # Client implementation for communicating with the LLM
│   └── server
│       └── server.go        # Web server entry point
├── go.mod                   # Module definition and dependencies
└── README.md                # Project documentation
```

## Features

- Web-based chat interface for interacting with the Ollama LLM.
- Stores ongoing chat history between the user and the LLM for the duration of the connection.
- Supplies the entire ongoing chat history as context to the LLM using the `messages` field in each request.
- Supports streaming responses from the LLM for improved interactivity.
- Allows selection of available models from a dropdown menu in the web UI.
- Displays the time taken for each prompt to generate a response.

## Setup Instructions

1. **Clone the repository:**
   ```sh
   git clone <repository-url>
   cd ollama-go-client-1
   ```

2. **Install dependencies:**
   ```sh
   go mod tidy
   ```

3. **Run the web server:**
   ```sh
   go run src/server/server.go
   ```

4. **Open the web interface:**
   - Visit `http://localhost:8080` in your browser.

## Usage

- Enter your prompt in the text area and select a model from the dropdown.
- Click "Generate" to send your prompt to the selected model.
- The response will appear below, along with the time taken to generate it.

## Notes

- The backend Go server listens on port **8080** and serves the web interface and API endpoints.
- The Ollama LLM backend must be running and accessible at `http://localhost:11434`.
- The legacy CLI interface is no longer included; use the web interface
