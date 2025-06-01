package main

// ModelInfo represents information about a model
type ModelInfo struct {
	ID            string  `json:"id"`
	Object        string  `json:"object"`
	Created       int64   `json:"created"`
	OwnedBy       string  `json:"owned_by"`
	MaxTokens     int     `json:"max_output_tokens"`
	ContextWindow int     `json:"context_window"`
	InputPrice    float64 `json:"input_price"`
	OutputPrice   float64 `json:"output_price"`
	CachingPrice  float64 `json:"caching_price"`
	CachedPrice   float64 `json:"cached_price"`
}

// ModelListResponse represents the response from the models API
type ModelListResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// TokenUsage represents the token counts for a request/response
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Cost represents the calculated cost for the request/response
type Cost struct {
	ModelInfo  ModelInfo `json:"model_info"`
	InputCost  float64   `json:"input_cost"`
	OutputCost float64   `json:"output_cost"`
	TotalCost  float64   `json:"total_cost"`
}

// AnthropicMessage represents a single message in the conversation
// @Description A message in the chat conversation
type AnthropicMessage struct {
	// The role of the message author (e.g., "user" or "assistant")
	// @Example user
	Role string `json:"role" example:"user"`
	// The content of the message
	// @Example Hello, how can you help me today?
	Content string `json:"content" example:"Hello, how can you help me today?"`
}

// AnthropicRequest represents the incoming request in Anthropic format
// @Description Request format for the chat API
type AnthropicRequest struct {
	// The model to use for completion
	// @Example gpt-4o-mini
	Model string `json:"model" example:"gpt-4o-mini"`
	// The messages to generate completions for
	Messages []AnthropicMessage `json:"messages"`
	// The maximum number of tokens to generate
	// @Example 1000
	MaxTokens int `json:"max_tokens,omitempty" example:"1000"`
	// Whether to stream the response
	// @Example false
	Stream bool `json:"stream,omitempty" example:"false"`
}

// AnthropicResponse represents the response in Anthropic format
// @Description Response format for the chat API
type AnthropicResponse struct {
	// The unique identifier for this completion
	// @Example msg_1234567890
	ID string `json:"id" example:"msg_1234567890"`
	// The type of the response
	// @Example message
	Type string `json:"type" example:"message"`
	// The role of the message author
	// @Example assistant
	Role string `json:"role" example:"assistant"`
	// The generated content
	// @Example I can help you with various tasks. What would you like to know?
	Content string `json:"content" example:"I can help you with various tasks. What would you like to know?"`
	// The model used for completion
	// @Example gpt-4o-mini
	Model string `json:"model" example:"gpt-4o-mini"`
	// The reason why the completion stopped
	// @Example stop
	StopReason string `json:"stop_reason" example:"stop"`
	// Token usage information
	Usage TokenUsage `json:"usage"`
	// Cost information
	Cost Cost `json:"cost"`
}

// AnthropicStreamResponse represents a streaming response in Anthropic format
// @Description Streaming response format for the chat API
type AnthropicStreamResponse struct {
	// The type of the stream event
	// @Example content_block_delta
	Type string `json:"type" example:"content_block_delta"`
	// The index of the content block
	// @Example 0
	Index int `json:"index" example:"0"`
	// The delta content
	Delta Delta `json:"delta"`
	// The reason why the stream stopped (if applicable)
	// @Example stop
	StopReason string `json:"stop_reason,omitempty" example:"stop"`
}

// Delta represents the incremental content in a streaming response
// @Description Incremental content in a streaming response
type Delta struct {
	// The content text
	// @Example Hello
	Text string `json:"text" example:"Hello"`
}

// ErrorResponse represents an error response
// @Description Error response format
type ErrorResponse struct {
	// The error message
	// @Example Invalid request body
	AnthropicMessage string `json:"message" example:"Invalid request body"`
	// The error code
	// @Example 400
	Code int `json:"code" example:"400"`
}

// OpenAI API models
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model     string          `json:"model"`
	Messages  []OpenAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens,omitempty"`
	Stream    bool            `json:"stream,omitempty"`
}

type OpenAIErrorMessage struct {
	AnthropicMessage string `json:"message"`
}

type OpenAIResponse struct {
	ID      string             `json:"id"`
	Error   OpenAIErrorMessage `json:"error"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Usage   TokenUsage         `json:"usage"`
	Choices []struct {
		AnthropicMessage OpenAIMessage `json:"message"`
		FinishReason     string        `json:"finish_reason"`
	} `json:"choices"`
}

type OpenAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta        OpenAIMessage `json:"delta"`
		FinishReason string        `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// HealthResponse represents the health check response
// @Description Health check response format
type HealthResponse struct {
	// Overall status of the service
	// @Example healthy
	Status string `json:"status" example:"healthy"`
	// Status of individual components
	Components map[string]ComponentHealth `json:"components"`
	// Timestamp of the health check
	// @Example 2024-03-20T15:04:05Z
	Time string `json:"time" example:"2024-03-20T15:04:05Z"`
}

// ComponentHealth represents the health status of a single component
// @Description Health status of a service component
type ComponentHealth struct {
	// Status of the component
	// @Example healthy
	Status string `json:"status" example:"healthy"`
	// Optional details about the component's health
	// @Example Connection timeout
	Details string `json:"details,omitempty" example:"Connection timeout"`
}
