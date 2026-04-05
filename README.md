# Polymarket Bot 🤖

A Go trading bot for [Polymarket](https://polymarket.com) — the world's largest prediction market.

**Prebuilt desktop app:** Installers and release assets: [GitHub Releases](https://github.com/Power-Howdy/polymarket-bot-go-wails/releases/latest).

## Features

- **Market discovery** via Gamma API (keyword filtering, volume sorting)
- **Order book data** via CLOB API with live bid/ask/spread
- **Three built-in strategies**: monitor, spread maker, momentum
- **Risk management**: exposure limits, per-market caps, stop-loss
- **Dry-run mode**: full simulation with no real orders
- **HMAC auth** for authenticated CLOB endpoints
- **Graceful shutdown** on SIGINT/SIGTERM

## Architecture

```
main.go
├── config/     — YAML config + env-var overrides
├── client/     — Polymarket API (Gamma + CLOB)
├── strategy/   — Monitor / Spread / Momentum strategies
├── risk/       — Exposure tracking, P&L, stop-loss
└── bot/        — Main loop, market selection, order execution
```

## Documentation

- **[Full guide (Polymarket basics + how to use the app)](docs/GUIDE.md)** — recommended read before live trading.
- In the **desktop app**, open the **Docs** sidebar tab for the same topics in the UI.

## Quick Start

### 1. Prerequisites
- Go 1.22+
- A Polygon wallet (for live trading)
- [Polymarket Builder Program](https://builders.polymarket.com) API keys (for trading)

### 2. Install dependencies
```bash
cd polymarket-bot
go mod tidy
```

### 3. Configure
```bash
cp config.example.yaml config.yaml
# Edit config.yaml — or set env vars (see below)
```

### 4. Run in dry-run mode (safe — no real orders)
```bash
go run . --dry-run
```

### 5. Build and run
```bash
go build -o polymarket-bot .
./polymarket-bot --config config.yaml
```

## Credentials via Environment Variables

```bash
export POLYMARKET_PRIVATE_KEY="your_hex_private_key"
export POLYMARKET_API_KEY="your_api_key"
export POLYMARKET_API_SECRET="your_api_secret"
export POLYMARKET_API_PASSPHRASE="your_passphrase"
```

## Strategies

### `monitor` (default)
Read-only. Polls markets, prints bid/ask/spread/volume. No orders placed. 
Good starting point to verify connectivity.

### `spread`
Posts limit orders on both sides of the mid-price.  
Configure `spread_bps` (width) and `order_size_usdc` (size per side).

```yaml
strategy:
  name: spread
  spread_bps: 50        # 0.5% each side
  order_size_usdc: 25.0
```

### `momentum`
Buys when price rises above `momentum_threshold`, sells when it falls.

```yaml
strategy:
  name: momentum
  momentum_threshold: 0.03   # 3% move triggers a trade
  momentum_lookback: 5       # compare current price to 5 ticks ago
  order_size_usdc: 20.0
```

## Risk Parameters

| Parameter | Description |
|-----------|-------------|
| `max_exposure_usdc` | Max total USDC across all open positions |
| `max_market_exposure_usdc` | Max USDC in a single market |
| `stop_loss_usdc` | Bot halts if realized P&L drops below this |

## API Reference

The bot uses two Polymarket APIs:

- **Gamma API** (`gamma-api.polymarket.com`) — market discovery, prices, metadata. Public, no auth required.
- **CLOB API** (`clob.polymarket.com`) — order book, order placement. Auth required for trading.

See [Polymarket Docs](https://docs.polymarket.com) for full API reference.

## ⚠️ Disclaimer

This bot is for educational purposes. Prediction market trading involves real financial risk. Always run in `--dry-run` mode first. Never trade more than you can afford to lose.
