# Polymarket Bot — User guide

This guide explains how to run this bot (desktop app) and the basics of trading on [Polymarket](https://polymarket.com). For API details, see the [official Polymarket documentation](https://docs.polymarket.com).

---

## What is Polymarket?

Polymarket is a **prediction market**: people trade contracts whose payout depends on real-world outcomes (elections, sports, crypto prices, geopolitics, and more).

- Each **market** asks a question with a defined resolution rule and end date.
- **Binary markets** (the most common in this bot) have two sides, often labeled **Yes** and **No**. You buy shares in one outcome.
- **Price as probability:** In a liquid binary market, a **Yes** share trading around **$0.60** roughly corresponds to the market implying a **~60%** chance that Yes resolves true. Prices move as traders update their views and as new information arrives.
- **Payout (simplified):** If you hold shares in the winning outcome when the market **resolves**, those shares typically settle toward **$1 USDC** per share; losing outcome shares go to **$0**. Actual mechanics depend on Polymarket’s rules and the specific market.
- **Collateral:** Trading uses **USDC** on **Polygon** (chain id **137** in this project). You need a wallet funded with USDC on Polygon for live trading, plus any Polymarket onboarding steps they require.

This bot does **not** explain tax, legal eligibility, or jurisdictional restrictions. Read Polymarket’s terms and only use products you are allowed to use.

---

## How this bot talks to Polymarket

| System | Role |
|--------|------|
| **Gamma API** | **Discovery** — lists markets, questions, volume, metadata. Public; no auth for read-only market data. |
| **CLOB API** | **Central limit order book** — order books (bids/asks), prices, and **placing/canceling orders**. Authenticated trading uses HMAC-signed requests with API keys from the [Builder Program](https://builders.polymarket.com). |

The bot discovers markets on Gamma, then pulls **order books** and (if not in dry-run) can submit orders on the CLOB.

---

## Running the desktop app (Wails)

1. **Install tooling:** [Go](https://go.dev/dl/) (see `go.mod` for version), [Node.js](https://nodejs.org/) for the frontend, and the [Wails CLI](https://wails.io/docs/gettingstarted/installation) if you build from source.
2. **Frontend assets:** From `frontend/`, run `npm install` and `npm run build` (or use `wails dev` for development).
3. **Build the app:** From the repo root, `wails build` produces a native binary for your OS.
4. **Configuration:** The app loads **JSON** via the file picker (same field names as `config.example.yaml`; convert YAML to JSON if you start from the example). Open **Settings**, browse to the file, adjust strategy and risk, then start the bot. You can also tune values in the UI after loading.

**Environment variables** (optional; override file values where implemented):

- `POLYMARKET_PRIVATE_KEY` — wallet private key (hex, often without `0x` — match what `config` expects)
- `POLYMARKET_API_KEY`, `POLYMARKET_API_SECRET`, `POLYMARKET_API_PASSPHRASE` — CLOB API credentials

Never commit secrets. Prefer env vars or a local config file ignored by git.

---

## Using the UI

| Area | Purpose |
|------|---------|
| **Dashboard** | High-level status, signals, risk summary while the bot runs. |
| **Markets** | Live view of markets the bot is tracking (bid/ask/spread/volume from polling). |
| **Positions** | Open / closed positions and P&amp;L tracking from the bot’s perspective. |
| **Backtest** | Run strategy logic on historical-style data (see in-app options). |
| **Settings** | Config file, API hosts, credentials, strategy parameters, risk limits, **dry run**, poll interval. |

- **Start / Stop** — Header controls; the bot loop runs in the background.
- **Dry run** — When enabled, the bot simulates or logs actions **without** sending real orders. Use this until you trust connectivity, market selection, and risk limits.

---

## Strategies (overview)

Configured under `strategy` in your config file (mirrored in Settings).

| Strategy | Behavior |
|----------|----------|
| **monitor** | Read-only. Polls markets and order books; **no orders**. Good for verifying APIs and watching spreads. |
| **spread** | Market-making style: places **limit** bids and asks around the mid-price. Parameters include **spread_bps** (width in basis points) and **order_size_usdc**. Requires live keys and sufficient balance; carries inventory and adverse-selection risk. |
| **momentum** | Reacts to short-term **price moves** vs a lookback window. Tunes: **momentum_threshold**, **momentum_lookback**, **order_size_usdc**. Can churn in noisy books; backtest and dry-run first. |

**Market filter:** `market_keywords` restricts discovery to questions containing those strings (case-insensitive in typical flows). Empty means “top markets by volume” up to **max_markets**.

---

## Risk settings

Under `risk` in config:

- **max_exposure_usdc** — Cap on total exposure the bot should allow across markets.
- **max_market_exposure_usdc** — Per-market cap.
- **stop_loss_usdc** — Stop trading when realized P&amp;L falls below this threshold (negative means max acceptable loss).

These are **bot-side guardrails**. They do not remove market risk, smart-contract risk, or API failure risk.

---

## Trading concepts (short glossary)

- **Bid / ask** — Best prices to sell vs buy in the book; **spread** is ask − bid.
- **Mid-price** — Rough fair value between bid and ask; used heuristically by strategies.
- **Limit order** — Price you specify; may not fill if the market moves away.
- **Liquidity** — Depth at each level; thin books mean larger **slippage** for your size.
- **Resolution** — When the oracle / rules determine the outcome; only then do winning shares pay out.

---

## Disclaimer

This software is for **education and research**. Prediction markets involve **financial risk**. Run **dry-run** first, use small sizes, and never trade more than you can afford to lose. The authors are not responsible for losses, bugs, or API changes.

---

## Further reading

- [Polymarket](https://polymarket.com)
- [Polymarket Docs](https://docs.polymarket.com)
- [Builder Program](https://builders.polymarket.com) (API keys)
