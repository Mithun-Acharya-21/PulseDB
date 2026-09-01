
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /pulsedb ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -s /bin/sh pulsedb
USER pulsedb

COPY --from=builder /pulsedb /usr/local/bin/pulsedb

EXPOSE 8080

ENTRYPOINT ["pulsedb"]
