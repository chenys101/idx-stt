
# Build Stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go module and sum files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 is important for creating a statically linked executable
# -o main specifies the output file name
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Final Stage
FROM alpine:latest

WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/main .
# Copy the config file
COPY config.yaml .
# Copy the database file
COPY data.db .
# Copy the frontend UI files
COPY stt/ui ./stt/ui

# Expose the port the application runs on
EXPOSE 8080

# The command to run the application
ENTRYPOINT ["./main"]
