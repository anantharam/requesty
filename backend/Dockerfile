# Use the official Golang image as base
FROM golang:1.22-alpine

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./main"] 