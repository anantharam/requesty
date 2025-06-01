package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"llm_gateway/logger"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "llm_gateway/docs"

	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           LLM Gateway API
// @version         1.0
// @description     A gateway service that converts between Anthropic and OpenAI API formats
// @host            localhost:8080
// @BasePath        /v1
// @schemes         http

var (
	// OPENAI_API_URL is the URL for the OpenAI API endpoint
	OPENAI_API_URL string
	// API_KEY is the authentication key for the OpenAI API
	API_KEY string
	// INFLUXDB_URL is the URL for the InfluxDB instance
	INFLUXDB_URL string
	// INFLUXDB_TOKEN is the authentication token for InfluxDB
	INFLUXDB_TOKEN string
	// PORT is the port number for the server to listen on
	PORT string
)

func init() {
	// Initialize configuration from environment variables with defaults
	OPENAI_API_URL = getEnv("OPENAI_API_URL", "https://router.requesty.ai/v1/chat/completions")
	API_KEY = getEnv("API_KEY", "sk-eyyhb/fzS1qMAhH15w/AaZPO/XZjAeAC3QVtP6VyE5eGpXlf2Q39LHOqkJ1YLpzK6HZ0MCo9ULMt8dQ5BzaGpupDNSDWmvvomsMVCEnlTQU=")
	if API_KEY == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	INFLUXDB_URL = getEnv("INFLUXDB_URL", "http://localhost:8086")
	INFLUXDB_TOKEN = getEnv("INFLUXDB_TOKEN", "")
	PORT = getEnv("PORT", "8080")
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func convertAnthropicToOpenAI(anthropicReq *AnthropicRequest) *OpenAIRequest {
	openaiMessages := make([]OpenAIMessage, len(anthropicReq.Messages))
	for i, msg := range anthropicReq.Messages {
		openaiMessages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return &OpenAIRequest{
		Model:     "openai/" + anthropicReq.Model,
		Messages:  openaiMessages,
		MaxTokens: anthropicReq.MaxTokens,
		Stream:    anthropicReq.Stream,
	}
}

func calculateCost(modelID string, usage TokenUsage) Cost {
	// Get model pricing from the API
	req, err := http.NewRequest("GET", "https://router.requesty.ai/v1/models", nil)
	if err != nil {
		log.Printf("Failed to create request for model pricing: %v", err)
		return Cost{}
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch model pricing: %v", err)
		return Cost{}
	}
	defer resp.Body.Close()

	var modelList ModelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelList); err != nil {
		log.Printf("Failed to decode model pricing: %v", err)
		return Cost{}
	}

	// Find the model modelInfo
	var modelInfo ModelInfo
	for _, model := range modelList.Data {
		if model.ID == modelID {
			modelInfo = model
			break
		}
	}

	// Calculate costs
	inputCost := float64(usage.PromptTokens) * modelInfo.InputPrice
	outputCost := float64(usage.CompletionTokens) * modelInfo.OutputPrice
	totalCost := inputCost + outputCost

	return Cost{
		ModelInfo:  modelInfo,
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  totalCost,
	}
}

func convertOpenAIToAnthropic(openaiResp *OpenAIResponse) *AnthropicResponse {
	if len(openaiResp.Choices) == 0 {
		return nil
	}

	choice := openaiResp.Choices[0]
	modelID := "openai/" + strings.TrimPrefix(openaiResp.Model, "openai/")
	cost := calculateCost(modelID, openaiResp.Usage)

	return &AnthropicResponse{
		ID:         openaiResp.ID,
		Type:       "message",
		Role:       choice.AnthropicMessage.Role,
		Content:    choice.AnthropicMessage.Content,
		Model:      openaiResp.Model,
		StopReason: choice.FinishReason,
		Usage:      openaiResp.Usage,
		Cost:       cost,
	}
}

func convertStreamResponse(openaiStream *OpenAIStreamResponse) *AnthropicStreamResponse {
	if len(openaiStream.Choices) == 0 {
		return nil
	}

	choice := openaiStream.Choices[0]
	return &AnthropicStreamResponse{
		Type:  "content_block_delta",
		Index: 0,
		Delta: Delta{
			Text: choice.Delta.Content,
		},
		StopReason: choice.FinishReason,
	}
}

// @Summary      Send messages to LLM
// @Description  Process chat messages through the LLM gateway
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        request body AnthropicRequest true "Chat request"
// @Success      200  {object}  AnthropicResponse
// @Success      200  {object}  AnthropicStreamResponse "When stream=true"
// @Failure      400  {object}  ErrorResponse
// @Failure      405  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /messages [post]
func handleMessages(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	startTime := time.Now()

	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.LogError(requestID, "method_not_allowed", "Invalid HTTP method")
		return
	}

	// Read and parse Anthropic request
	var anthropicReq AnthropicRequest
	if err := json.NewDecoder(r.Body).Decode(&anthropicReq); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		logger.LogError(requestID, "invalid_request", "Invalid request body")
		return
	}

	// Log the incoming request
	err := logger.LogRequest(requestID, anthropicReq.Model, len(anthropicReq.Messages), anthropicReq.Stream)
	if err != nil {
		log.Printf("Failed to log request: %v", err)
	}

	// Convert to OpenAI format
	openaiReq := convertAnthropicToOpenAI(&anthropicReq)

	// Forward to OpenAI API
	openaiReqBody, err := json.Marshal(openaiReq)
	if err != nil {
		sendErrorResponse(w, "Error preparing request", http.StatusInternalServerError)
		logger.LogError(requestID, "marshal_error", "Error preparing request")
		return
	}

	// Create request with API key
	req, err := http.NewRequest(http.MethodPost, OPENAI_API_URL, bytes.NewBuffer(openaiReqBody))
	if err != nil {
		sendErrorResponse(w, "Error creating request", http.StatusInternalServerError)
		logger.LogError(requestID, "request_creation_error", "Error creating request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+API_KEY)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sendErrorResponse(w, "Error forwarding request", http.StatusInternalServerError)
		logger.LogError(requestID, "api_error", "Error forwarding request")
		return
	}
	defer resp.Body.Close()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Handle streaming response
	if anthropicReq.Stream {
		handleStreamingResponse(w, resp, requestID)
		return
	}

	// Handle regular response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sendErrorResponse(w, "Error reading response", http.StatusInternalServerError)
		logger.LogError(requestID, "response_read_error", "Error reading response")
		return
	}

	// Parse OpenAI response
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		sendErrorResponse(w, "Error parsing response", http.StatusInternalServerError)
		logger.LogError(requestID, "response_parse_error", "Error parsing response")
		return
	}

	// Convert to Anthropic format
	anthropicResp := convertOpenAIToAnthropic(&openaiResp)
	if anthropicResp == nil {
		sendErrorResponse(w, "Invalid response from OpenAI", http.StatusInternalServerError)
		logger.LogError(requestID, "conversion_error", "Invalid response from OpenAI")
		return
	}

	// Log the response
	responseTime := time.Since(startTime)
	err = logger.LogResponse(requestID, anthropicReq.Model, responseTime, http.StatusOK, false)
	if err != nil {
		log.Printf("Failed to log response: %v", err)
	}

	// Send response
	json.NewEncoder(w).Encode(anthropicResp)
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		AnthropicMessage: message,
		Code:             code,
	})
}

func handleStreamingResponse(w http.ResponseWriter, resp *http.Response, requestID string) {
	// Set up streaming response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.LogError(requestID, "streaming_unsupported", "Streaming unsupported")
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	reader := bufio.NewReader(resp.Body)
	chunkNumber := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.LogError(requestID, "stream_read_error", fmt.Sprintf("Error reading stream: %v", err))
			log.Printf("Error reading stream: %v", err)
			break
		}

		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove "data: " prefix
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			break
		}

		// Parse OpenAI stream response
		var openaiStream OpenAIStreamResponse
		if err := json.Unmarshal([]byte(data), &openaiStream); err != nil {
			logger.LogError(requestID, "stream_parse_error", fmt.Sprintf("Error parsing stream data: %v", err))
			log.Printf("Error parsing stream data: %v", err)
			continue
		}

		// Convert to Anthropic format
		anthropicStream := convertStreamResponse(&openaiStream)
		if anthropicStream == nil || strings.TrimSpace(anthropicStream.Delta.Text) == "" {
			continue
		}

		// Log streaming chunk
		chunkNumber++
		err = logger.LogStreamingChunk(requestID, len(anthropicStream.Delta.Text), chunkNumber)
		if err != nil {
			log.Printf("Failed to log streaming chunk: %v", err)
		}

		// Send the converted response
		if err := json.NewEncoder(w).Encode(anthropicStream); err != nil {
			logger.LogError(requestID, "stream_write_error", fmt.Sprintf("Error encoding stream response: %v", err))
			log.Printf("Error encoding stream response: %v", err)
			continue
		}
		fmt.Fprint(w, "\n")
		flusher.Flush()
	}
}

// @Summary      Health check endpoint
// @Description  Check the health of the service and its dependencies
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  ErrorResponse
// @Router       /health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	health := HealthResponse{
		Status: "healthy",
		Components: map[string]ComponentHealth{
			"server": {
				Status: "healthy",
			},
			"influxdb": {
				Status: "healthy",
			},
		},
		Time: time.Now().Format(time.RFC3339),
	}

	// Check InfluxDB health
	if err := logger.CheckHealth(); err != nil {
		health.Status = "unhealthy"
		health.Components["influxdb"] = ComponentHealth{
			Status:  "unhealthy",
			Details: err.Error(),
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// Convert component health to string map for logging
	componentStatuses := make(map[string]string)
	for name, component := range health.Components {
		componentStatuses[name] = component.Status
		if component.Details != "" {
			componentStatuses[name+"_details"] = component.Details
		}
	}

	// Log health check
	duration := time.Since(startTime)
	if err := logger.LogHealthCheck(health.Status, componentStatuses, duration); err != nil {
		log.Printf("Failed to log health check: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func main() {
	// Initialize logger
	if err := logger.Initialize(INFLUXDB_URL, INFLUXDB_TOKEN); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Register API routes
	http.HandleFunc("/v1/messages", handleMessages)
	http.HandleFunc("/v1/health", handleHealth)

	// Serve Swagger UI
	http.HandleFunc("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to generated swagger file
	))

	serverAddr := ":" + PORT
	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", PORT)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
