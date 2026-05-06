#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════
# PUYAD Kalkanı — Kali Linux / Docker Kurulum Scripti
# ═══════════════════════════════════════════════════════════
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

banner() {
    echo -e "${CYAN}"
    echo "  ╔══════════════════════════════════════════════╗"
    echo "  ║     🛡️  PUYAD Kalkanı — Linux Hardening     ║"
    echo "  ║        Docker Kurulum & Başlatma             ║"
    echo "  ╚══════════════════════════════════════════════╝"
    echo -e "${NC}"
}

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; }

banner

# ── 1. Docker kurulu mu? ──
if ! command -v docker &>/dev/null; then
    warn "Docker bulunamadı. Kuruluyor..."
    
    # Kali/Debian/Ubuntu otomatik kurulum
    sudo apt-get update -qq
    sudo apt-get install -y -qq \
        ca-certificates curl gnupg lsb-release

    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/debian/gpg | \
        sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg

    # Kali, Debian tabanlı olduğu için bookworm kullan
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
      https://download.docker.com/linux/debian bookworm stable" | \
      sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    sudo apt-get update -qq
    sudo apt-get install -y -qq \
        docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    sudo systemctl enable --now docker
    sudo usermod -aG docker "$USER"
    log "Docker kuruldu!"
else
    log "Docker zaten kurulu: $(docker --version)"
fi

# ── 2. Docker Compose var mı? ──
if docker compose version &>/dev/null; then
    log "Docker Compose (plugin): $(docker compose version --short)"
    COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    log "Docker Compose (standalone): $(docker-compose --version)"
    COMPOSE="docker-compose"
else
    warn "Docker Compose bulunamadı. Plugin kuruluyor..."
    sudo apt-get install -y -qq docker-compose-plugin
    COMPOSE="docker compose"
fi

# ── 3. Build & Çalıştır ──
echo ""
log "Docker imajı build ediliyor..."
$COMPOSE build --no-cache

echo ""
log "Container başlatılıyor..."
$COMPOSE up -d kalkani

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════${NC}"
log "PUYAD Kalkanı çalışıyor!"
echo ""
echo -e "  🌐 Dashboard:  ${GREEN}http://localhost:8080${NC}"
echo -e "  📡 API Scan:    ${GREEN}http://localhost:8080/api/scan${NC}"
echo -e "  📋 Sonuçlar:    ${GREEN}http://localhost:8080/api/results${NC}"
echo ""
echo -e "  ${YELLOW}Host sistemini taramak için:${NC}"
echo -e "  ${CYAN}$COMPOSE --profile host-scan up -d${NC}"
echo -e "  → http://localhost:8081"
echo ""
echo -e "  ${YELLOW}Logları izle:${NC}  $COMPOSE logs -f kalkani"
echo -e "  ${YELLOW}Durdur:${NC}        $COMPOSE down"
echo -e "${CYAN}═══════════════════════════════════════════════${NC}"
