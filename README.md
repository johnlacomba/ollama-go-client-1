# Ollama Go Client

This project is a Go client for communicating with a local Ollama LLM (Large Language Model). It provides a simple interface to send prompts and receive responses from the model.

## Project Structure

```
ollama-go-client
├── src
│   ├── main.go          # Entry point of the application
│   └── client
│       └── client.go    # Client implementation for communicating with the LLM
├── go.mod               # Module definition and dependencies
└── README.md            # Project documentation
```

## Features

- Stores ongoing chat history between the user and the LLM for the duration of the connection.
- Supplies the entire ongoing chat history as context to the LLM using the `messages` field in each request.
- Supports streaming responses from the LLM for improved interactivity.

## Setup Instructions

1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd ollama-go-client
   ```

2. **Install dependencies:**
   ```
   go mod tidy
   ```

3. **Build the project:**
   ```
   go build -o ollama-go-client src/main.go
   ```

4. **Run the application:**
   ```
   ./ollama-go-client
   ```

## Usage

After running the application, you can input prompts to communicate with the Ollama LLM. The client will send the prompts to the model and display the responses.
