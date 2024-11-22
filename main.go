package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

var apiKey = os.Getenv("OPENAI_API_KEY")

func main() {
	if apiKey == "" {
		apiKey = "YOUR_API_KEY"
	}

	// Custom handler for root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// Serve static files for other paths
			http.ServeFile(w, r, "./public"+r.URL.Path)
			return
		}
		// Serve context.html for the root path
		http.ServeFile(w, r, "./public/context.html")
	})

	// API route
	http.HandleFunc("/api/identify-area", identifyAreaHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func identifyAreaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Prompt      string  `json:"prompt"`
		AIModel     string  `json:"aiModel"`
		Temperature float64 `json:"temperature"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &reqBody); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	model := reqBody.AIModel
	if model == "" {
		model = "gpt-4o-mini" // Default model
	}

	temperature := reqBody.Temperature

	client := openai.NewClient(apiKey)

	// Create the completion request
	resp, err := client.CreateChatCompletion(r.Context(), openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: reqBody.Prompt,
			},
		},
		Temperature: float32(temperature),
		MaxTokens:   150, // Adjust max tokens as needed
	})

	if err != nil {
		log.Printf("Error processing the request: %v\n", err)
		http.Error(w, "An error occurred while processing your request.", http.StatusInternalServerError)
		return
	}

	result := ""
	if len(resp.Choices) > 0 {
		result = resp.Choices[0].Message.Content
	}

	log.Printf("ChatGPT Response: %s\n", result)

	response := map[string]string{"areaOfInterest": result}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
