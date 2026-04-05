import { useState, useCallback } from 'react'
import { loadConfig, selectConfigFile } from '../lib/wails'

const DEFAULT_CFG = {
  clob_host: 'https://clob.polymarket.com',
  gamma_host: 'https://gamma-api.polymarket.com',
  api_key: '', api_secret: '', api_passphrase: '', private_key: '',
  chain_id: 137,
  dry_run: true,
  poll_interval: '10s',
  strategy: {
    name: 'momentum',
    market_keywords: [],
    max_markets: 5,
    spread_bps: 50,
    order_size_usdc: 10,
    momentum_threshold: 0.03,
    momentum_lookback: 5,
  },
  risk: {
    max_exposure_usdc: 500,
    max_market_exposure_usdc: 100,
    stop_loss_usdc: -50,
  },
}

export function useConfig() {
  const [cfg, setCfg] = useState(DEFAULT_CFG)
  const [configPath, setConfigPath] = useState('')
  const [loadErr, setLoadErr] = useState(null)

  const pickAndLoad = useCallback(async () => {
    try {
      const path = await selectConfigFile()
      if (!path) return
      const loaded = await loadConfig(path)
      setConfigPath(path)
      setCfg(loaded)
      setLoadErr(null)
    } catch (e) {
      setLoadErr(e?.message ?? String(e))
    }
  }, [])

  const update = useCallback((patch) => {
    setCfg(prev => ({ ...prev, ...patch }))
  }, [])

  const updateStrategy = useCallback((patch) => {
    setCfg(prev => ({ ...prev, strategy: { ...prev.strategy, ...patch } }))
  }, [])

  const updateRisk = useCallback((patch) => {
    setCfg(prev => ({ ...prev, risk: { ...prev.risk, ...patch } }))
  }, [])

  return { cfg, configPath, loadErr, pickAndLoad, update, updateStrategy, updateRisk }
}
