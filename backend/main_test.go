package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleNonStreamingMessages(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		request        AnthropicRequest
		expectedStatus int
		validateResp   func(*testing.T, *AnthropicResponse)
	}{
		{
			name: "Basic conversation test",
			request: AnthropicRequest{
				Model: "gpt-4o-mini",
				Messages: []AnthropicMessage{
					{
						Role:    "user",
						Content: "What is the capital of France?",
					},
				},
				MaxTokens: 100,
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, resp *AnthropicResponse) {
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, "message", resp.Type)
				assert.NotEmpty(t, resp.Content)
				assert.Contains(t, resp.Content, "Paris", "Response should mention Paris")
				assert.NotEmpty(t, resp.Model)
				assert.NotEmpty(t, resp.StopReason)
			},
		},
		{
			name: "Technical question test",
			request: AnthropicRequest{
				Model: "gpt-4o-mini",
				Messages: []AnthropicMessage{
					{
						Role:    "user",
						Content: "Explain what a REST API is.",
					},
				},
				MaxTokens: 150,
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, resp *AnthropicResponse) {
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, "message", resp.Type)
				assert.NotEmpty(t, resp.Content)
				assert.Contains(t, resp.Content, "REST", "Response should mention REST")
				assert.NotEmpty(t, resp.Model)
				assert.NotEmpty(t, resp.StopReason)
			},
		},
		{
			name: "Test Request without Model",
			request: AnthropicRequest{
				Messages: []AnthropicMessage{
					{
						Role:    "user",
						Content: "Explain what a REST API is.",
					},
				},
				MaxTokens: 150,
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp:   nil,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request body
			reqBody, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle the request
			handleMessages(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If expecting success, validate response
			if tc.expectedStatus == http.StatusOK {
				var resp AnthropicResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				assert.NoError(t, err)

				// Run custom validation
				tc.validateResp(t, &resp)
			}
		})
	}
}

func TestHandleMessagesStreaming(t *testing.T) {
	tests := []struct {
		name           string
		request        AnthropicRequest
		expectedStatus int
		validateStream func(*testing.T, *bufio.Scanner)
	}{
		{
			name: "Basic streaming test",
			request: AnthropicRequest{
				Model: "gpt-4o-mini",
				Messages: []AnthropicMessage{
					{
						Role:    "user",
						Content: "Count from 1 to 3.",
					},
				},
				MaxTokens: 100,
				Stream:    true,
			},
			expectedStatus: http.StatusOK,
			validateStream: func(t *testing.T, scanner *bufio.Scanner) {
				var messageCount int
				var foundDone bool

				for scanner.Scan() {
					line := scanner.Text()
					if line == "" {
						continue
					}

					if line == "data: [DONE]" {
						foundDone = true
						break
					}

					var streamResp AnthropicStreamResponse
					err := json.Unmarshal([]byte(line), &streamResp)
					assert.NoError(t, err)

					// Validate stream response structure
					assert.Equal(t, "content_block_delta", streamResp.Type)
					assert.Equal(t, 0, streamResp.Index)
					assert.NotEmpty(t, streamResp.Delta.Text)

					messageCount++
				}

				// Ensure we got multiple messages and the [DONE] marker
				assert.True(t, messageCount > 0, "Should receive multiple stream messages")
				assert.True(t, foundDone, "Should receive [DONE] marker")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request body
			reqBody, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle the request
			handleMessages(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// For streaming responses, validate the stream
			if tc.expectedStatus == http.StatusOK {
				scanner := bufio.NewScanner(rr.Body)
				tc.validateStream(t, scanner)
			}
		})
	}
}

func TestHandleMessagesInvalidMethod(t *testing.T) {
	// Create GET request (invalid method)
	req := httptest.NewRequest(http.MethodGet, "/v1/messages", nil)
	rr := httptest.NewRecorder()

	// Handle the request
	handleMessages(rr, req)

	// Check response
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleMessagesInvalidBody(t *testing.T) {
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Handle the request
	handleMessages(rr, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestConvertStreamResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    OpenAIStreamResponse
		expected *AnthropicStreamResponse
	}{
		{
			name: "Valid stream response",
			input: OpenAIStreamResponse{
				ID:      "test-id",
				Object:  "chat.completion.chunk",
				Created: 1234567890,
				Model:   "claude-2",
				Choices: []struct {
					Delta        OpenAIMessage `json:"delta"`
					FinishReason string        `json:"finish_reason,omitempty"`
				}{
					{
						Delta: OpenAIMessage{
							Content: "Hello",
						},
					},
				},
			},
			expected: &AnthropicStreamResponse{
				Type:  "content_block_delta",
				Index: 0,
				Delta: Delta{
					Text: "Hello",
				},
			},
		},
		{
			name: "Empty choices",
			input: OpenAIStreamResponse{
				ID: "test-id",
				Choices: []struct {
					Delta        OpenAIMessage `json:"delta"`
					FinishReason string        `json:"finish_reason,omitempty"`
				}{},
			},
			expected: nil,
		},
		{
			name: "With finish reason",
			input: OpenAIStreamResponse{
				ID: "test-id",
				Choices: []struct {
					Delta        OpenAIMessage `json:"delta"`
					FinishReason string        `json:"finish_reason,omitempty"`
				}{
					{
						Delta: OpenAIMessage{
							Content: "Bye",
						},
						FinishReason: "stop",
					},
				},
			},
			expected: &AnthropicStreamResponse{
				Type:  "content_block_delta",
				Index: 0,
				Delta: Delta{
					Text: "Bye",
				},
				StopReason: "stop",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertStreamResponse(&tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
