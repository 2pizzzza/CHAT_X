# Use a Golang base image
FROM golang:1.22 AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o sso ./cmd/chat_x/main.go

# Use a lightweight base image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary file from the builder stage
COPY --from=builder /app/sso ./

COPY .env ./

# Command to run the application
CMD ["./chat_x"]
