# ═══════════════════════════════════════════════════════════
# PUYAD Kalkanı — Multi-stage Docker Build
# ═══════════════════════════════════════════════════════════

# ── Stage 1: Frontend Build ──
FROM node:22-alpine AS frontend-builder
WORKDIR /build/frontend
COPY frontend/package.json ./
COPY frontend/package-lock.json* ./
RUN npm install --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# ── Stage 2: Go Backend Build ──
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /kalkani ./cmd/kalkani/

# ── Stage 3: Final Runtime Image ──
FROM debian:bookworm-slim

# Güvenlik araçları
RUN apt-get update && apt-get install -y --no-install-recommends \
    openssh-server \
    ufw \
    iptables \
    iproute2 \
    procps \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Dizin yapısı
RUN mkdir -p /opt/kalkani/configs /opt/kalkani/frontend/dist /var/lib/kalkani/backups

# Backend binary
COPY --from=backend-builder /kalkani /opt/kalkani/kalkani
RUN chmod +x /opt/kalkani/kalkani

# Frontend dist
COPY --from=frontend-builder /build/frontend/dist/ /opt/kalkani/frontend/dist/

# Config dosyaları
COPY configs/ /opt/kalkani/configs/

# Çalışma dizini
WORKDIR /opt/kalkani

# Port
EXPOSE 8080

# Sağlık kontrolü
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/results || exit 1

# Başlat
CMD ["./kalkani"]
