package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"ollama-go-client/src/client"
)

func main() {
	endpoint := "http://localhost:11434" // Adjust the endpoint as necessary
	timeout := 120 * time.Second

	ollamaClient := client.NewClient(endpoint, timeout)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your prompt for the Ollama LLM (type 'exit' to quit):")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "exit" {
			break
		}

		response, err := ollamaClient.SendRequest(input)
		if err != nil {
			log.Println("Error sending request:", err)
			continue
		}

		fmt.Println("Response:", response.Message.Content)
	}

	fmt.Println("Exiting the Ollama LLM client.")
	os.Exit(0)
}
