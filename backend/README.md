# LLM Gateway Service

A gateway service that converts between Anthropic and OpenAI API formats.

## Environment Variables

The service can be configured using the following environment variables:

### Required Variables
- `API_KEY`: Your OpenAI API key (Required)

### Optional Variables with Defaults
- `OPENAI_API_URL`: OpenAI API endpoint (Default: "https://router.requesty.ai/v1/chat/completions")
- `INFLUXDB_URL`: InfluxDB instance URL (Default: "http://localhost:8086")
- `INFLUXDB_TOKEN`: InfluxDB authentication token (Default: "my-super-secret-admin-token")
- `PORT`: Server port number (Default: "8080")

## API Documentation

The service provides a Swagger UI interface for exploring and testing the API endpoints. Once the service is running, you can access the Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

The API provides the following endpoints:

- `POST /v1/messages`: Send messages to the LLM
  - Supports both regular and streaming responses
  - Request format follows Anthropic's API structure
  - Returns responses in Anthropic's format

## Running the Service

1. Set up your environment variables:
   ```bash
   export API_KEY=your-api-key-here
   # Optionally set other variables
   export PORT=3000
   ```

2. Run the service:
   ```bash
   go run main.go
   ```

The service will start and listen on the configured port (default: 8080).

## Development

To regenerate the Swagger documentation after making changes to the API:

```bash
swag init
```

This will update the Swagger files in the `docs` directory. 