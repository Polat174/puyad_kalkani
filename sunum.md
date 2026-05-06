# PUYAD Kalkanı — Proje Sunumu

*(Bu doküman, teknik odaklı 2 sayfalık sunum formatına göre hazırlanmıştır.)*

---

## SAYFA 1: Proje Özeti ve Mimari Altyapı

**Proje Adı:** PUYAD Kalkanı — Otomatik Linux Hardening ve Denetim Sistemi
**Amaç:** Kali Linux başta olmak üzere Linux sistemlerinin CIS, STIG ve yerel güvenlik standartlarına (Harun ŞEKER Sıkılaştırma Yönergeleri) göre otomatik analiz edilmesi ve zafiyetlerin "tek tıkla" giderilmesi.

**Teknik Mimari:**
* **Backend (Go):** Yüksek performanslı ve eşzamanlı (concurrent) tarama motoru. 
  * `chi/v5` Router altyapısı ve `Server-Sent Events (SSE)` ile frontend'e canlı veri akışı.
  * Modüler arayüzler: *Scanner* (Denetçi) ve *Fixer* (İyileştirici).
* **Frontend (React + TS):** Modern, karanlık tema (dark-mode) odaklı ve dinamik durum yönetimli (Vite tabanlı) kontrol paneli.
* **Altyapı (Docker):** Host makineyi izole bir konteynerdan yönetebilmek için `privileged` mod, host-network entegrasyonu ve root filesystem mount (`/:/host`) kullanan multi-stage mimari.

**Temel Yetenekler:**
* **100+ Güvenlik Parametresi:** SSH konfigürasyonları, Kernel (Sysctl) TCP/IP sıkılaştırmaları, Dosya ve Dizin Yetkileri (SUID/SGID/World-writable), Parola Politikaları (PAM pwquality), Firewall (UFW/Iptables).
* **Güvenlik Sigortası:** Yapılan tüm sistem değişiklikleri otomatik olarak yedeklenir (Backup/Restore mekanizması).

---

## SAYFA 2: Teknik Çözümler ve Demo Akışı

**Öne Çıkan Teknik Geliştirmeler:**
1. **Host İzolasyonunu Aşma:** Sistem konteyner (Docker) içerisinde çalışmasına rağmen, host makinenin `/etc` ve `/proc` gibi kritik noktalarına read-write yetkisiyle müdahale edebilmektedir.
2. **Asenkron Canlı İzleme:** Tarama sırasında çalışan ağır bash komutları (örn: `find / -xdev`), EventSource (SSE) üzerinden frontend'e anlık "Taranıyor..." logları basarak kullanıcı deneyimini kopukluklardan kurtarır. (Race-condition korumalıdır).
3. **Güvenli Yürütme (Context Timeouts):** Sistem komutlarına (`exec.Command`) context-timeout eklenerek zombi proses oluşumu ve deadlock durumları engellenmiştir.

**Canlı Demo Senaryosu (Kısa ve Net):**
1. **Başlatma:** Sistemin Kali Linux üzerinde `docker compose up` ile ayağa kaldırılması ve Dashboard'a erişim (Port 8080/8081).
2. **Analiz:** Sistemin mevcut güvenlik puanının hesaplanması ve anlık tarama metriklerinin ekranda canlı izlenmesi.
3. **Aksiyon:** Kritik risk taşıyan kırmızı uyarıların (Örn: *SSH Root Login aktif*, *SUID dosyaları tespit edildi*) arayüz üzerinden "Düzelt" butonu ile kapatılması.
4. **Doğrulama:** Sistemin arka planında (`cat /etc/ssh/sshd_config` vb.) komutlarla sorunun gerçekten çözüldüğünün canlı ispatı.
