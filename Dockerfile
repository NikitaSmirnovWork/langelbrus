FROM golang:1.21-alpine AS builder

# Need gcc for go-sqlite3
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./main.go

# ── Runtime ──
FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app
COPY --from=builder /app/server .
COPY index.html .

EXPOSE 8080
ENV GIN_MODE=release
ENV DB_PATH=/data/wordbot.db

# Create data directory for SQLite persistence
RUN mkdir -p /data

CMD ["./server"]
