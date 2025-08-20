# Use official Go image
FROM golang:1.24 AS builder

# Set working directory
WORKDIR /app

# Copy Go files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go binary
RUN go build -o server

# Create a smaller image for deployment
FROM gcr.io/distroless/base-debian12

WORKDIR /
COPY --from=builder /app/server /

# Expose port
EXPOSE 8080

# Run the app
CMD ["/server"]
