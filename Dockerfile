# Multi-stage build: build → distroless runtime
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download deps first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build both binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/worker ./cmd/worker

# Runtime image — distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/worker /app/worker
COPY --from=builder /app/config/scoring_weights.yaml /app/config/scoring_weights.yaml

# Default entrypoint is the API server.
# Override CMD to run the worker: docker run <image> /app/worker
ENTRYPOINT ["/app/api"]
