import { useState, useEffect, useCallback, useRef } from 'react'
import { getState, isRunning, startBot, stopBot, onStateUpdate, onBotError } from '../lib/wails'

const EMPTY = {
  markets: [], positions: [], recentSignals: [],
  risk: { totalExposure: 0, realizedPnL: 0, tradesTotal: 0, tradesBuys: 0, tradesSells: 0 },
  uptime: '0s', lastUpdate: '—', strategy: '—', dryRun: false, errors: [],
}

export function useBotState() {
  const [state, setState] = useState(EMPTY)
  const [running, setRunning] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const pollRef = useRef(null)

  // Subscribe to push events from the Go backend
  useEffect(() => {
    const unsub = onStateUpdate((snap) => {
      setState(snap)
    })
    const unsubErr = onBotError((msg) => {
      setError(msg)
      setRunning(false)
    })
    return () => { unsub(); unsubErr() }
  }, [])

  // Fallback: poll every 2 s when events aren't wired (dev mode)
  useEffect(() => {
    pollRef.current = setInterval(async () => {
      if (!running) return
      try {
        const snap = await getState()
        if (snap) setState(snap)
      } catch (_) {}
    }, 2000)
    return () => clearInterval(pollRef.current)
  }, [running])

  const start = useCallback(async (cfg) => {
    setLoading(true)
    setError(null)
    try {
      await startBot(cfg)
      setRunning(true)
    } catch (e) {
      setError(e?.message ?? String(e))
    } finally {
      setLoading(false)
    }
  }, [])

  const stop = useCallback(async () => {
    await stopBot()
    setRunning(false)
  }, [])

  // Sync running state on mount
  useEffect(() => {
    isRunning().then(setRunning).catch(() => {})
  }, [])

  return { state, running, loading, error, start, stop }
}
