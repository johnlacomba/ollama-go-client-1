# Ollama Go Client

This project is a Go client for communicating with a local Ollama LLM (Large Language Model). It provides a web interface with two modes: a standard "User-To-Model" chat and a "Model-To-Model" conversational interface with advanced features for managing complex AI conversations.

## Project Structure

```
ollama-go-client-1
├── src
│   ├── static
│   │   ├── index.html       # Web UI for both chat modes
│   │   └── style.css        # Dark theme styling
│   ├── ollamaAPIWrapper
│   │   └── ollamaAPIWrapper.go # Wrapper for Ollama API calls
│   └── server
│       └── server.go        # HTTP server and API endpoint logic
├── go.mod
└── README.md
```

## Features

### General
- **Dark Theme UI**: Modern dark interface optimized for extended use
- **Dual Interfaces**: Switch between direct user-model chat and automated model-to-model conversations
- **Real-Time Streaming**: Responses stream token-by-token using Server-Sent Events (SSE) for immediate feedback
- **Dynamic Model Loading**: Automatically fetches available models from your local Ollama instance
- **Session Management**: Server maintains separate chat histories per user session
- **Timestamped Messages**: All messages display generation timestamps
- **Performance Metrics**: Each response shows accurate generation duration including model load time
- **Smart Auto-Scroll**: Chat windows auto-scroll only when at bottom, preserving your reading position
- **Tabbed Interface**: Clean navigation between User-To-Model and Model-To-Model modes

### User-To-Model Interface
- **Standard Chat**: Familiar chat interface for direct interaction with any Ollama model
- **Image Support**: Paste images directly into prompts for multi-modal conversations
- **Persistent Prompts**: Set system-level instructions that are automatically prepended to every message
- **Advanced Parameters**: Fine-tune model behavior with temperature, top-p, frequency penalty, and presence penalty controls
- **Enter Key Submission**: Press Enter to send messages (Shift+Enter for new lines)
- **Auto-Expanding Input**: Text area automatically resizes based on content

### Model-To-Model Interface
- **Automated Conversations**: Create conversations between two different models
- **Individual Model Configuration**: Set different parameters and persistent prompts for each model
- **Interactive Prompt Injection**: Insert new prompts mid-conversation to steer the discussion
- **Real-Time Control**: Start, stop, and modify conversations at any time
- **Intelligent History Management**: Uses conversation summarization to maintain context while preventing prompt dilution
- **Persistent Prompt Prioritization**: Each model's core instructions are emphasized over conversation history
- **Pending Prompt System**: Injected prompts show as "pending" and seamlessly integrate when the current response completes

### Advanced Conversation Management
- **Context Summarization**: For M2M conversations, the system automatically summarizes the last 10 messages to maintain coherent context while preventing models from being influenced by the other model's persistent prompts
- **Structured Prompting**: Persistent prompts are clearly marked as "primary instructions" to ensure models follow their intended roles
- **Clean Session Boundaries**: Each new M2M conversation starts with a fresh context

## How Streaming Works

1. Browser sends `POST` request to `/api/chat` with model, prompt, options, and optional image data
2. Server processes persistent prompts and conversation history according to the interface mode
3. For M2M mode, server generates conversation summaries when needed to maintain context clarity
4. Server forwards structured request to Ollama's `/api/chat` endpoint with `"stream": true`
5. Server streams each token chunk back to client as SSE `data:` events with JSON payload
6. Final `event: done` message includes accurate generation duration from Ollama
7. Client receives events, accumulates tokens, and updates UI in real-time with proper message structure

## API Endpoints

- `GET /`  
  Serves the main application and static assets

- `POST /api/chat`  
  Initiates streaming chat response with advanced context management
  - **Request Body**: 
    ```json
    {
      "model": "<model_name>",
      "prompt": "<user_input>", 
      "persistent_prompt": "<system_instructions>",
      "image": "<base64_image_data>",
      "summarize_history": true,
      "options": {
        "temperature": 0.8,
        "top_p": 0.95,
        "frequency_penalty": 0.8,
        "presence_penalty": 0.6
      }
    }
    ```
  - **Response Stream**:
    - `data: {"token":"<text_chunk>"}`
    - `event: done\ndata: {"duration":"<total_duration>"}`

- `POST /api/clear-history`  
  Clears chat history for current session (used when starting new M2M conversations)

- `GET /api/tags`  
  Returns JSON array of available Ollama models

## Setup

```sh
# Ensure dependencies are available
go mod tidy

# Start the server
go run src/server/server.go
```

Visit: `http://localhost:8080`

**Requirements**: Local Ollama instance running on `http://localhost:11434`

## Usage Examples

### User-To-Model
1. Set a persistent prompt like "You are a helpful coding assistant. Provide concise, well-commented code examples."
2. Adjust generation parameters using the sliders
3. Type your question and press Enter or click Send
4. Paste images directly into the input for visual questions

### Model-To-Model
1. Select two different models (e.g., `llama2` and `codellama`)
2. Set distinct persistent prompts:
   - Model A: "You are a creative storyteller. Write engaging narratives."
   - Model B: "You are a critical editor. Provide constructive feedback on stories."
3. Enter initial prompt: "Write a short story about time travel"
4. Click "Insert Prompt" to start the conversation
5. Inject new prompts anytime: "Now focus on the character development"
6. Use "Stop" to end the conversation

## Advanced Features

### Context Management
- **History Summarization**: M2M conversations automatically summarize context to prevent prompt confusion
- **Persistent Prompt Emphasis**: System instructions are clearly prioritized over conversation history
- **Clean Message Structure**: Proper DOM structure prevents spacing issues and ensures clean presentation

### User Experience
- **Responsive Design**: Interface adapts to different screen sizes
- **Keyboard Shortcuts**: Enter for submission, Shift+Enter for line breaks
- **Visual Feedback**: Blinking cursors during generation, pending prompt indicators
- **Session Persistence**: Chat history maintained during browser session

## Customization

### Server Configuration
- Modify `endpoint` in `server.go` to use different Ollama instances
- Adjust `timeout` values for different response time requirements
- Configure summarization logic in `getSummary` function

### UI Customization
- Edit `style.css` for theme modifications
- Adjust slider ranges and default values in `index.html`
- Modify message rendering functions for different display formats

## Technical Notes

- **Session Management**: Based on remote IP address (simple approach for development)
- **History Persistence**: In-memory only; restarting server clears all sessions  
- **Model Loading**: Automatic detection of available Ollama models
- **Error Handling**: Graceful fallbacks for network issues and model errors
- **Performance**: Efficient streaming with minimal client-side buffering

## Requirements

- Go 1.18+
- Local Ollama installation with at least one model downloaded
- Modern web browser with SSE support
- Network access to localhost:11434 (default Ollama port)
