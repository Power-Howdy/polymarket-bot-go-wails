/**
 * Wails bridge — wraps Go-bound methods and event subscriptions.
 * Falls back to mock data when running outside the Wails shell (e.g. `npm run dev`).
 */

const IS_WAILS = typeof window !== 'undefined' && !!window.go

// ── Go method calls ───────────────────────────────────────────────────────────

export async function loadConfig(path) {
  if (IS_WAILS) return window.go.app.App.LoadConfig(path)
  return mockConfig()
}

export async function startBot(cfg) {
  if (IS_WAILS) return window.go.app.App.StartBot(cfg)
  console.info('[mock] startBot', cfg)
}

export async function stopBot() {
  if (IS_WAILS) return window.go.app.App.StopBot()
  console.info('[mock] stopBot')
}

export async function isRunning() {
  if (IS_WAILS) return window.go.app.App.IsRunning()
  return false
}

export async function getState() {
  if (IS_WAILS) return window.go.app.App.GetState()
  return mockState()
}

export async function runBacktest(req) {
  if (IS_WAILS) return window.go.app.App.RunBacktest(req)
  return mockBacktest(req)
}

export async function selectConfigFile() {
  if (IS_WAILS) return window.go.app.App.SelectConfigFile()
  return 'config.json'
}

// ── Event subscriptions ───────────────────────────────────────────────────────

export function onStateUpdate(cb) {
  if (IS_WAILS) {
    window.runtime.EventsOn('state:update', cb)
    return () => window.runtime.EventsOff('state:update')
  }
  return () => {}
}

export function onBotError(cb) {
  if (IS_WAILS) {
    window.runtime.EventsOn('bot:error', cb)
    return () => window.runtime.EventsOff('bot:error')
  }
  return () => {}
}

// ── Mock data for dev mode ────────────────────────────────────────────────────

function mockConfig() {
  return {
    clob_host: 'https://clob.polymarket.com',
    gamma_host: 'https://gamma-api.polymarket.com',
    dry_run: true,
    strategy: { name: 'momentum', max_markets: 5, order_size_usdc: 10, spread_bps: 50, momentum_threshold: 0.03, momentum_lookback: 5 },
    risk: { max_exposure_usdc: 500, max_market_exposure_usdc: 100, stop_loss_usdc: -50 },
    poll_interval: '10s',
  }
}

function mockState() {
  return {
    markets: [
      { question: 'Will BTC exceed $100k by Dec 2025?', outcome: 'Yes', bid: 0.42, ask: 0.44, mid: 0.43, spread: 0.02, volume: 125430, signal: 'BUY @$0.431' },
      { question: 'Will ETH hit $5k in 2025?', outcome: 'Yes', bid: 0.31, ask: 0.33, mid: 0.32, spread: 0.02, volume: 88210, signal: 'NONE' },
      { question: 'Will the Fed cut rates in Q1 2025?', outcome: 'Yes', bid: 0.67, ask: 0.69, mid: 0.68, spread: 0.02, volume: 312000, signal: 'SELL @$0.688' },
    ],
    positions: [
      { tokenId: 'tok1', market: 'Will BTC exceed $100k by Dec 2025?', outcome: 'Yes', totalSize: 23.2, avgBuyPrice: 0.41, realizedPnL: 4.2, unrealized: 2.1 },
    ],
    recentSignals: [
      { side: 'BUY',  market: 'Will BTC exceed $100k...', price: 0.431, size: 23.2, reason: 'momentum UP 3.2% over 5 ticks' },
      { side: 'NONE', market: 'Will ETH hit $5k...',      price: 0.32,  size: 0,    reason: 'change=0.5% below threshold' },
      { side: 'SELL', market: 'Will the Fed cut rates...', price: 0.688, size: 14.5, reason: 'momentum DOWN 4.1% over 5 ticks' },
    ],
    risk: { totalExposure: 241.2, realizedPnL: 4.2, tradesTotal: 7, tradesBuys: 5, tradesSells: 2 },
    uptime: '5m32s',
    lastUpdate: '14:22:01',
    strategy: 'momentum',
    dryRun: true,
    errors: [],
  }
}

function mockBacktest(req) {
  const curve = Array.from({ length: 60 }, (_, i) => (Math.random() - 0.48) * 0.5 * (i + 1))
  for (let i = 1; i < curve.length; i++) curve[i] += curve[i - 1]
  return {
    strategy: req.strategy,
    pnl: curve[curve.length - 1],
    trades: Math.floor(req.ticks / 25),
    wins: Math.floor(req.ticks / 45),
    losses: Math.floor(req.ticks / 70),
    winRate: 0.54,
    maxDrawdown: -12.4,
    equityCurve: curve,
    fills: Array.from({ length: 8 }, (_, i) => ({
      side: i % 2 === 0 ? 'BUY' : 'SELL',
      price: 0.4 + Math.random() * 0.2,
      size: 20 + Math.random() * 10,
      cost: 10,
    })),
  }
}
