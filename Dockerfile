# Build stage
FROM golang:1.24-alpine@sha256:738d1cf061836894ff6bb8c33881080ac66de8cf0586615012a0c8f592649cfa AS builder

WORKDIR /src

# Download dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build statically linked binary
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags="-s -w" -o /bin/chrome-tab-cleaner ./cmd/chrome-tab-cleaner

# Runtime stage
FROM alpine:3.21@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

# Install ca-certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user and home directory
RUN addgroup -S chrometabcleaner && adduser -S chrometabcleaner -G chrometabcleaner && \
    mkdir -p /home/chrometabcleaner && \
    chown -R chrometabcleaner:chrometabcleaner /home/chrometabcleaner

COPY --from=builder /bin/chrome-tab-cleaner /usr/local/bin/chrome-tab-cleaner

USER chrometabcleaner
WORKDIR /home/chrometabcleaner

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q -O- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/chrome-tab-cleaner"]
CMD ["-serve-healthz", ":8080"]
