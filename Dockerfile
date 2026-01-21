# Build Go binary
FROM golang:1.22-alpine AS go-builder

WORKDIR /build

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o feedback ./cmd/feedback

# Build Tailwind CSS
FROM node:20-alpine AS css-builder

WORKDIR /build

# Copy package files
COPY package.json ./
RUN npm install

# Copy Tailwind config and templates
COPY tailwind.config.js ./
COPY web ./web

# Build CSS
RUN npx tailwindcss -i web/static/css/input.css -o web/static/css/output.css --minify

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from go-builder
COPY --from=go-builder /build/feedback .

# Copy static assets
COPY --from=css-builder /build/web/static ./web/static
COPY web/templates ./web/templates

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set data directory to volume
VOLUME ["/data"]

# Run the application
CMD ["./feedback"]
