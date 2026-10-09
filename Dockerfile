# The build stage
# Cross-compile on the host arch so `--platform linux/amd64` builds stay fast on Apple Silicon
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app

# Download dependencies first so they are cached until go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o api ./cmd/api

# The run stage
FROM scratch
WORKDIR /app
# CA certificates are needed for outgoing HTTPS calls (SendGrid)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/api .

# Run as an unprivileged user (nobody)
USER 65534:65534
EXPOSE 8080
CMD ["./api"]
