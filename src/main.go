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
	timeout := 300 * time.Second

	ollamaClient := client.NewClient(endpoint, timeout)

	models, err := ollamaClient.ListModels()
	if err != nil {
		log.Println("Error listing models:", err)
		return
	}

	fmt.Println("Available models:")
	for i, model := range models {
		fmt.Printf("%d. %s\n", i+1, model)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter the number of the model you want to use (or 'exit' to quit):")
modelLoop:
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "exit" {
			break
		}

		var model int
		_, err := fmt.Sscan(input, &model)
		if err != nil || model < 1 || model > len(models) {
			fmt.Println("Invalid model number. Please try again.")
			continue
		}

		selectedModel := models[model-1]

		fmt.Println("Enter your prompt for the Ollama LLM (type 'exit' to quit):")
	promptLoop:
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break promptLoop
			}

			input := scanner.Text()
			if input == "exit" {
				break modelLoop // Break out of the outer loop
			}

			temperature := 0.7 // Set default temperature
			topP := 0.95       // Set default top_p

			response, err := ollamaClient.SendRequest(selectedModel, input, float32(temperature), float32(topP), 0.0, 0.0, 0, 0, 0, 0)
			if err != nil {
				log.Println("Error sending request:", err)
				continue
			}

			fmt.Println("Response:", response.Message.Content)
		}
	}
	fmt.Println("Exiting the Ollama LLM client.")
	os.Exit(0)
}
