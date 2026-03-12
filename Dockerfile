# syntax=docker/dockerfile:1

# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# modernc.org/sqlite is pure Go — no CGO required
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dns-filter .

# ── Runtime stage ──────────────────────────────────────────────────────────────
FROM alpine:3.21

# ca-certificates for upstream DNS-over-HTTPS if ever needed
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/dns-filter ./dns-filter

VOLUME ["/data"]

# 53  — DNS (UDP + TCP)
# 9090 — HTTP management API
EXPOSE 53/udp 53/tcp 9090/tcp

ENTRYPOINT ["./dns-filter"]
CMD ["-db", "/data/dns.db", "-api", ":9090"]
