#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════
# PUYAD Kalkanı — Manuel Doğrulama Scripti
# Dashboard sonuçlarını terminal'den doğrular
# Kullanım: sudo ./scripts/test.sh
# ═══════════════════════════════════════════════════════════
set -uo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

PASS=0; FAIL=0; WARN=0

check() {
    local desc="$1" expected="$2" actual="$3"
    if [[ "$actual" == "$expected" ]]; then
        echo -e "  ${GREEN}✅ PASS${NC}  $desc  →  ${BOLD}$actual${NC}"
        ((PASS++))
    else
        echo -e "  ${RED}❌ FAIL${NC}  $desc  →  mevcut: ${RED}$actual${NC}  beklenen: ${GREEN}$expected${NC}"
        ((FAIL++))
    fi
}

warn_check() {
    local desc="$1" actual="$2"
    echo -e "  ${YELLOW}⚠️  WARN${NC}  $desc  →  ${BOLD}$actual${NC}"
    ((WARN++))
}

separator() {
    echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

echo -e "${CYAN}"
echo "  ╔══════════════════════════════════════════════╗"
echo "  ║   🛡️  PUYAD Kalkanı — Manuel Doğrulama      ║"
echo "  ║   Dashboard sonuçlarını terminal'den kontrol ║"
echo "  ╚══════════════════════════════════════════════╝"
echo -e "${NC}"

# ═══════════════════════════════════════
# 1. SSH KONTROLLERI
# ═══════════════════════════════════════
separator "1. SSH GÜVENLİĞİ (/etc/ssh/sshd_config)"

SSHD="/etc/ssh/sshd_config"
if [[ -f "$SSHD" ]]; then
    get_ssh() {
        grep -i "^$1" "$SSHD" 2>/dev/null | awk '{print $2}' | head -1
    }

    check "PermitRootLogin"        "no"  "$(get_ssh PermitRootLogin)"
    check "PasswordAuthentication" "no"  "$(get_ssh PasswordAuthentication)"
    check "PermitEmptyPasswords"   "no"  "$(get_ssh PermitEmptyPasswords)"
    check "MaxAuthTries"           "4"   "$(get_ssh MaxAuthTries)"
    check "X11Forwarding"          "no"  "$(get_ssh X11Forwarding)"
    check "AllowAgentForwarding"   "no"  "$(get_ssh AllowAgentForwarding)"
    check "ClientAliveInterval"    "300" "$(get_ssh ClientAliveInterval)"
    check "ClientAliveCountMax"    "2"   "$(get_ssh ClientAliveCountMax)"
    check "LoginGraceTime"         "60"  "$(get_ssh LoginGraceTime)"
    check "MaxSessions"            "4"   "$(get_ssh MaxSessions)"
    check "Protocol"               "2"   "$(get_ssh Protocol)"
    check "UsePAM"                 "yes" "$(get_ssh UsePAM)"
    check "Banner"                 "/etc/issue.net" "$(get_ssh Banner)"
else
    warn_check "sshd_config bulunamadı" "$SSHD"
fi

# ═══════════════════════════════════════
# 2. KERNEL / SYSCTL KONTROLLERI
# ═══════════════════════════════════════
separator "2. KERNEL / SYSCTL PARAMETRELERİ"

get_sysctl() {
    sysctl -n "$1" 2>/dev/null || cat "/proc/sys/$(echo "$1" | tr '.' '/')" 2>/dev/null || echo "(okunamadı)"
}

check "net.ipv4.ip_forward"                        "0" "$(get_sysctl net.ipv4.ip_forward)"
check "net.ipv4.conf.all.accept_redirects"          "0" "$(get_sysctl net.ipv4.conf.all.accept_redirects)"
check "net.ipv4.conf.default.accept_redirects"      "0" "$(get_sysctl net.ipv4.conf.default.accept_redirects)"
check "net.ipv4.conf.all.send_redirects"            "0" "$(get_sysctl net.ipv4.conf.all.send_redirects)"
check "net.ipv4.conf.default.send_redirects"        "0" "$(get_sysctl net.ipv4.conf.default.send_redirects)"
check "net.ipv4.conf.all.accept_source_route"       "0" "$(get_sysctl net.ipv4.conf.all.accept_source_route)"
check "net.ipv4.conf.default.accept_source_route"   "0" "$(get_sysctl net.ipv4.conf.default.accept_source_route)"
check "net.ipv4.conf.all.log_martians"              "1" "$(get_sysctl net.ipv4.conf.all.log_martians)"
check "net.ipv4.conf.default.log_martians"          "1" "$(get_sysctl net.ipv4.conf.default.log_martians)"
check "net.ipv4.tcp_syncookies"                     "1" "$(get_sysctl net.ipv4.tcp_syncookies)"
check "net.ipv4.icmp_echo_ignore_broadcasts"        "1" "$(get_sysctl net.ipv4.icmp_echo_ignore_broadcasts)"
check "net.ipv4.icmp_ignore_bogus_error_responses"  "1" "$(get_sysctl net.ipv4.icmp_ignore_bogus_error_responses)"
check "net.ipv4.conf.all.rp_filter"                 "1" "$(get_sysctl net.ipv4.conf.all.rp_filter)"
check "net.ipv4.conf.default.rp_filter"             "1" "$(get_sysctl net.ipv4.conf.default.rp_filter)"
check "net.ipv6.conf.all.accept_redirects"          "0" "$(get_sysctl net.ipv6.conf.all.accept_redirects)"
check "net.ipv6.conf.default.accept_redirects"      "0" "$(get_sysctl net.ipv6.conf.default.accept_redirects)"
check "net.ipv6.conf.all.accept_source_route"       "0" "$(get_sysctl net.ipv6.conf.all.accept_source_route)"
check "kernel.randomize_va_space (ASLR)"            "2" "$(get_sysctl kernel.randomize_va_space)"
check "kernel.dmesg_restrict"                       "1" "$(get_sysctl kernel.dmesg_restrict)"
check "kernel.kptr_restrict"                        "1" "$(get_sysctl kernel.kptr_restrict)"
check "kernel.yama.ptrace_scope"                    "1" "$(get_sysctl kernel.yama.ptrace_scope)"
check "fs.suid_dumpable"                            "0" "$(get_sysctl fs.suid_dumpable)"
check "fs.protected_hardlinks"                      "1" "$(get_sysctl fs.protected_hardlinks)"
check "fs.protected_symlinks"                       "1" "$(get_sysctl fs.protected_symlinks)"

# ═══════════════════════════════════════
# 3. FIREWALL KONTROLLERI
# ═══════════════════════════════════════
separator "3. GÜVENLİK DUVARI"

if command -v ufw &>/dev/null; then
    UFW_STATUS=$(sudo ufw status 2>/dev/null | head -1)
    if echo "$UFW_STATUS" | grep -qi "active"; then
        check "UFW durumu" "active" "active"
    else
        check "UFW durumu" "active" "inactive"
    fi
else
    warn_check "UFW kurulu değil" "(bulunamadı)"
fi

if command -v iptables &>/dev/null; then
    INPUT_POLICY=$(sudo iptables -L INPUT -n 2>/dev/null | head -1 | grep -oP 'policy \K\w+')
    FORWARD_POLICY=$(sudo iptables -L FORWARD -n 2>/dev/null | head -1 | grep -oP 'policy \K\w+')
    check "iptables INPUT politikası"   "DROP" "${INPUT_POLICY:-bilinmiyor}"
    check "iptables FORWARD politikası" "DROP" "${FORWARD_POLICY:-bilinmiyor}"
else
    warn_check "iptables bulunamadı" "(kurulu değil)"
fi

# ═══════════════════════════════════════
# 4. KARA LİSTE KONTROLLERI
# ═══════════════════════════════════════
separator "4. KARA LİSTE (Tehlikeli Paketler)"

BLACKLIST=(telnetd telnet-server rsh-server rsh rlogin talk talk-server
           tftp-server tftp vsftpd xinetd ypserv ypbind rpcbind
           avahi-daemon cups nfs-kernel-server nfs-common slapd snmpd samba squid)

for pkg in "${BLACKLIST[@]}"; do
    if dpkg -l "$pkg" 2>/dev/null | grep -q "^ii"; then
        check "Paket: $pkg" "yüklü değil" "YÜKLÜ"
    else
        check "Paket: $pkg" "yüklü değil" "yüklü değil"
    fi
done

# ═══════════════════════════════════════
# 5. KULLANICI GÜVENLİĞİ
# ═══════════════════════════════════════
separator "5. KULLANICI GÜVENLİĞİ"

# UID 0 kontrolü
UID0_USERS=$(awk -F: '$3==0 && $1!="root" {print $1}' /etc/passwd)
if [[ -z "$UID0_USERS" ]]; then
    check "UID 0 (root dışı)" "sadece root" "sadece root"
else
    check "UID 0 (root dışı)" "sadece root" "$UID0_USERS"
fi

# Boş parola kontrolü
if [[ -r /etc/shadow ]]; then
    EMPTY_PW=$(awk -F: '$2=="" {print $1}' /etc/shadow | tr '\n' ', ')
    if [[ -z "$EMPTY_PW" ]]; then
        check "Boş parolalı hesap" "yok" "yok"
    else
        check "Boş parolalı hesap" "yok" "$EMPTY_PW"
    fi
else
    warn_check "shadow dosyası okunamadı (root gerekli)" ""
fi

# Umask kontrolü
CURRENT_UMASK=$(umask)
check "Umask" "0027" "$CURRENT_UMASK"

# ═══════════════════════════════════════
# 6. API KARŞILAŞTIRMASI
# ═══════════════════════════════════════
separator "6. API SONUÇLARI İLE KARŞILAŞTIRMA"

echo -e "  ${CYAN}Dashboard API'den sonuçlar çekiliyor...${NC}"
if command -v curl &>/dev/null; then
    API_RESULT=$(curl -s http://localhost:8080/api/results 2>/dev/null)
    if [[ -n "$API_RESULT" && "$API_RESULT" != *"error"* ]]; then
        API_PASS=$(echo "$API_RESULT" | grep -o '"pass":[0-9]*' | cut -d: -f2)
        API_FAIL=$(echo "$API_RESULT" | grep -o '"fail":[0-9]*' | cut -d: -f2)
        API_WARN=$(echo "$API_RESULT" | grep -o '"warn":[0-9]*' | cut -d: -f2)
        API_SCORE=$(echo "$API_RESULT" | grep -o '"score":[0-9]*' | cut -d: -f2)
        echo -e "  ${BOLD}Dashboard:${NC}  Pass=${GREEN}$API_PASS${NC}  Fail=${RED}$API_FAIL${NC}  Warn=${YELLOW}$API_WARN${NC}  Skor=${CYAN}$API_SCORE${NC}"
    else
        warn_check "API'ye bağlanılamadı" "önce tarama başlatın"
    fi
else
    warn_check "curl bulunamadı" ""
fi

# ═══════════════════════════════════════
# SONUÇ
# ═══════════════════════════════════════
echo ""
echo -e "${CYAN}═══════════════════════════════════════════════${NC}"
echo -e "${BOLD}  SONUÇ${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════${NC}"
echo -e "  ${BOLD}Manuel Test:${NC}  Pass=${GREEN}$PASS${NC}  Fail=${RED}$FAIL${NC}  Warn=${YELLOW}$WARN${NC}"
TOTAL=$((PASS + FAIL))
if [[ $TOTAL -gt 0 ]]; then
    SCORE=$(( PASS * 100 / TOTAL ))
    echo -e "  ${BOLD}Güvenlik Puanı:${NC} ${CYAN}$SCORE / 100${NC}"
fi
echo ""
echo -e "  ${YELLOW}→ Bu sonuçları Dashboard ile karşılaştırın.${NC}"
echo -e "  ${YELLOW}  Aynıysa sistem gerçek değerleri tarıyor demektir.${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════${NC}"
