# ---------- Build stage ----------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /pulsedb ./cmd/api

# ---------- Runtime stage ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /pulsedb /usr/local/bin/pulsedb

EXPOSE 8080

ENTRYPOINT ["pulsedb"]
