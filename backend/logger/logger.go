package logger

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var (
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
)

const (
	bucket = "llm_metrics"
	org    = "my-org"
)

// Initialize sets up the InfluxDB client
func Initialize(url, token string) error {
	client = influxdb2.NewClient(url, token)

	// Create a blocking write client
	writeAPI = client.WriteAPIBlocking(org, bucket)

	// Test connection
	_, err := client.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect to InfluxDB: %v", err)
	}

	return nil
}

// Close releases resources used by the logger
func Close() {
	if client != nil {
		client.Close()
	}
}

// CheckHealth verifies the connection to InfluxDB
func CheckHealth() error {
	if client == nil {
		return fmt.Errorf("InfluxDB client not initialized")
	}

	// Try to ping InfluxDB with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Health(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to InfluxDB: %v", err)
	}

	return nil
}

// LogRequest logs the incoming request metadata
func LogRequest(requestID string, model string, messageCount int, streamEnabled bool) error {
	point := influxdb2.NewPoint(
		"llm_requests",
		map[string]string{
			"request_id": requestID,
			"model":      model,
		},
		map[string]interface{}{
			"message_count": messageCount,
			"streaming":     streamEnabled,
		},
		time.Now(),
	)

	return writeAPI.WritePoint(context.Background(), point)
}

// LogResponse logs the response metadata
func LogResponse(requestID string, model string, responseTime time.Duration, status int, errorOccurred bool) error {
	point := influxdb2.NewPoint(
		"llm_responses",
		map[string]string{
			"request_id": requestID,
			"model":      model,
		},
		map[string]interface{}{
			"response_time_ms": responseTime.Milliseconds(),
			"status":           status,
			"error":            errorOccurred,
		},
		time.Now(),
	)

	return writeAPI.WritePoint(context.Background(), point)
}

// LogStreamingChunk logs information about each streaming chunk
func LogStreamingChunk(requestID string, chunkSize int, chunkNumber int) error {
	point := influxdb2.NewPoint(
		"llm_streaming_chunks",
		map[string]string{
			"request_id": requestID,
		},
		map[string]interface{}{
			"chunk_size":   chunkSize,
			"chunk_number": chunkNumber,
		},
		time.Now(),
	)

	return writeAPI.WritePoint(context.Background(), point)
}

// LogError logs error events
func LogError(requestID string, errorType string, errorMessage string) error {
	point := influxdb2.NewPoint(
		"llm_errors",
		map[string]string{
			"request_id": requestID,
			"type":       errorType,
		},
		map[string]interface{}{
			"message": errorMessage,
		},
		time.Now(),
	)

	return writeAPI.WritePoint(context.Background(), point)
}

// LogHealthCheck logs information about health check requests and responses
func LogHealthCheck(status string, components map[string]string, duration time.Duration) error {
	p := influxdb2.NewPoint(
		"health_checks",
		map[string]string{
			"status": status,
		},
		map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
			"components":  components,
		},
		time.Now(),
	)

	return writeAPI.WritePoint(context.Background(), p)
}
