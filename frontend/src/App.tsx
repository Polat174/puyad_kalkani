import { useState, useEffect, useCallback, useRef } from 'react'
import './App.css'

interface ScanResult {
  id: string
  category: string
  description: string
  status: 'PASS' | 'FAIL' | 'WARN' | 'INFO'
  current_value: string
  expected_value: string
  fix_command?: string
}

interface ScanReport {
  timestamp: string
  hostname: string
  results: ScanResult[]
  summary: {
    total: number
    pass: number
    fail: number
    warn: number
    info: number
    score: number
  }
}

interface FixResult {
  id: string
  success: boolean
  message: string
}

const API = import.meta.env.DEV ? 'http://localhost:8080' : ''

function App() {
  const [report, setReport] = useState<ScanReport | null>(null)
  const [scanning, setScanning] = useState(false)
  const [fixing, setFixing] = useState<string | null>(null)
  const [fixingAll, setFixingAll] = useState(false)
  const [filter, setFilter] = useState<string>('all')
  const [liveResults, setLiveResults] = useState<ScanResult[]>([])
  const [currentScanner, setCurrentScanner] = useState('')
  const [toasts, setToasts] = useState<{id: number; msg: string; type: string}[]>([])
  const toastId = useRef(0)

  const addToast = useCallback((msg: string, type: string) => {
    const id = ++toastId.current
    setToasts(t => [...t, { id, msg, type }])
    setTimeout(() => setToasts(t => t.filter(x => x.id !== id)), 4000)
  }, [])

  const startScan = useCallback(async () => {
    setScanning(true)
    setLiveResults([])
    setReport(null)
    setCurrentScanner('')

    // SSE dinlemeye başla
    const es = new EventSource(`${API}/api/events`)
    es.addEventListener('scan_progress', (e) => {
      const d = JSON.parse(e.data)
      setCurrentScanner(d.scanner + ' — ' + d.status)
    })
    es.addEventListener('scan_result', (e) => {
      const r: ScanResult = JSON.parse(e.data)
      setLiveResults(prev => [...prev, r])
    })
    es.addEventListener('scan_complete', () => {
      es.close()
    })

    // Bağlantının kurulmasını bekle (race condition önleme)
    await new Promise<void>(resolve => {
      es.onopen = () => resolve()
      setTimeout(resolve, 1000) // Fallback timeout
    })

    try {
      const res = await fetch(`${API}/api/scan`)
      const data: ScanReport = await res.json()
      setReport(data)
    } catch (err) {
      addToast('Tarama başarısız: sunucuya bağlanılamadı', 'error')
    } finally {
      setScanning(false)
      setCurrentScanner('')
    }
  }, [addToast])

  const fixItem = useCallback(async (id: string) => {
    setFixing(id)
    try {
      const res = await fetch(`${API}/api/fix/${id}`, { method: 'POST' })
      const data: FixResult = await res.json()
      if (data.success) {
        addToast(`${id}: düzeltildi ✓`, 'success')
      } else {
        addToast(`${id}: ${data.message}`, 'error')
      }
    } catch {
      addToast('Düzeltme başarısız', 'error')
    } finally {
      setFixing(null)
    }
  }, [addToast])

  const fixAll = useCallback(async () => {
    setFixingAll(true)
    try {
      const res = await fetch(`${API}/api/fix-all`, { method: 'POST' })
      const data: FixResult[] = await res.json()
      const ok = data.filter(r => r.success).length
      const fail = data.filter(r => !r.success).length
      addToast(`Toplu düzeltme: ${ok} başarılı, ${fail} başarısız`, ok > 0 ? 'success' : 'error')
    } catch {
      addToast('Toplu düzeltme başarısız', 'error')
    } finally {
      setFixingAll(false)
    }
  }, [addToast])

  // İlk yüklemede son sonuçları getir
  useEffect(() => {
    fetch(`${API}/api/results`).then(r => r.json()).then(setReport).catch(() => {})
  }, [])

  const results = report?.results || liveResults
  const filteredResults = filter === 'all' ? results : results.filter(r => r.status === filter)
  const summary = report?.summary || { total: 0, pass: 0, fail: 0, warn: 0, info: 0, score: 0 }

  return (
    <div className="app">
      {/* Toast notifications */}
      <div className="toast-container">
        {toasts.map(t => (
          <div key={t.id} className={`toast toast-${t.type}`}>{t.msg}</div>
        ))}
      </div>

      {/* Header */}
      <header className="header">
        <div className="header-content">
          <div className="brand">
            <div className="shield-icon">🛡️</div>
            <div>
              <h1>PUYAD Kalkanı</h1>
              <p className="subtitle">Linux Hardening Tarama Sistemi</p>
            </div>
          </div>
          <div className="header-actions">
            {report && (
              <button className="btn btn-secondary" onClick={fixAll} disabled={fixingAll || summary.fail === 0}>
                {fixingAll ? '⏳ Düzeltiliyor...' : `🔧 Tümünü Düzelt (${summary.fail})`}
              </button>
            )}
            <button className="btn btn-primary" onClick={startScan} disabled={scanning}>
              {scanning ? '⏳ Taranıyor...' : '🔍 Taramayı Başlat'}
            </button>
          </div>
        </div>
      </header>

      <main className="main">
        {/* Live scanning indicator */}
        {scanning && (
          <div className="scan-live">
            <div className="scan-live-bar" />
            <div className="scan-live-text">
              <span className="pulse-dot" />
              {currentScanner || 'Tarama başlatılıyor...'}
            </div>
          </div>
        )}

        {/* Summary Cards */}
        {(report || liveResults.length > 0) && (
          <section className="summary-grid">
            <div className="card card-score">
              <div className="card-value">{summary.score}</div>
              <div className="card-label">Güvenlik Puanı</div>
              <div className="score-ring">
                <svg viewBox="0 0 36 36">
                  <path className="ring-bg" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                  <path className="ring-fg" strokeDasharray={`${summary.score}, 100`} d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                </svg>
              </div>
            </div>
            <div className="card card-pass" onClick={() => setFilter('PASS')}>
              <div className="card-value">{summary.pass}</div>
              <div className="card-label">Geçti</div>
              <div className="card-icon">✅</div>
            </div>
            <div className="card card-fail" onClick={() => setFilter('FAIL')}>
              <div className="card-value">{summary.fail}</div>
              <div className="card-label">Başarısız</div>
              <div className="card-icon">❌</div>
            </div>
            <div className="card card-warn" onClick={() => setFilter('WARN')}>
              <div className="card-value">{summary.warn}</div>
              <div className="card-label">Uyarı</div>
              <div className="card-icon">⚠️</div>
            </div>
            <div className="card card-total" onClick={() => setFilter('all')}>
              <div className="card-value">{summary.total}</div>
              <div className="card-label">Toplam</div>
              <div className="card-icon">📊</div>
            </div>
          </section>
        )}

        {/* Filter bar */}
        {results.length > 0 && (
          <div className="filter-bar">
            <span className="filter-label">Filtre:</span>
            {['all', 'PASS', 'FAIL', 'WARN'].map(f => (
              <button key={f} className={`filter-btn ${filter === f ? 'active' : ''}`} onClick={() => setFilter(f)}>
                {f === 'all' ? 'Tümü' : f === 'PASS' ? '✅ Geçti' : f === 'FAIL' ? '❌ Başarısız' : '⚠️ Uyarı'}
              </button>
            ))}
            {report && <span className="result-info">{report.hostname} • {new Date(report.timestamp).toLocaleString('tr-TR')}</span>}
          </div>
        )}

        {/* Results Table */}
        {filteredResults.length > 0 && (
          <div className="table-wrap">
            <table className="results-table">
              <thead>
                <tr>
                  <th>Durum</th>
                  <th>Kategori</th>
                  <th>Açıklama</th>
                  <th>Mevcut Değer</th>
                  <th>Beklenen Değer</th>
                  <th>İşlem</th>
                </tr>
              </thead>
              <tbody>
                {filteredResults.map((r, i) => (
                  <tr key={r.id || i} className={`row-${r.status.toLowerCase()}`}>
                    <td>
                      <span className={`badge badge-${r.status.toLowerCase()}`}>
                        {r.status === 'PASS' ? '✅' : r.status === 'FAIL' ? '❌' : r.status === 'WARN' ? '⚠️' : 'ℹ️'} {r.status}
                      </span>
                    </td>
                    <td><span className="category-tag">{r.category}</span></td>
                    <td>{r.description}</td>
                    <td><code>{r.current_value}</code></td>
                    <td><code>{r.expected_value}</code></td>
                    <td>
                      {r.status === 'FAIL' && r.fix_command && (
                        <button className="btn btn-fix" onClick={() => fixItem(r.id)} disabled={fixing === r.id}>
                          {fixing === r.id ? '⏳' : '🔧'} Düzelt
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Empty state */}
        {!scanning && results.length === 0 && (
          <div className="empty-state">
            <div className="empty-icon">🛡️</div>
            <h2>PUYAD Kalkanı</h2>
            <p>Linux sisteminizi CIS Benchmark standartlarına göre tarayın</p>
            <p className="empty-features">SSH • Kernel/Sysctl • Firewall • Servisler • Kullanıcılar</p>
            <button className="btn btn-primary btn-lg" onClick={startScan}>
              🔍 Taramayı Başlat
            </button>
          </div>
        )}
      </main>

      <footer className="footer">
        <p>PUYAD Kalkanı v1.0 — Linux Hardening Tarama &amp; Düzeltme Sistemi</p>
      </footer>
    </div>
  )
}

export default App
