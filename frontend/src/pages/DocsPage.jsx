import React from 'react'
import styles from './DocsPage.module.css'

export default function DocsPage() {
  return (
    <div className={styles.page}>
      <p className={styles.intro}>
        How to use this app and core ideas behind Polymarket trading. For the same content as a repo file, see{' '}
        <span className="mono">docs/GUIDE.md</span>. Official API reference:{' '}
        <a className={styles.link} href="https://docs.polymarket.com" target="_blank" rel="noreferrer">
          docs.polymarket.com
        </a>
        .
      </p>

      <section className={styles.section}>
        <h2 className={styles.h2}>What is Polymarket?</h2>
        <p className={styles.p}>
          Polymarket is a <strong>prediction market</strong>: traders buy and sell contracts tied to whether future
          events happen (politics, sports, crypto, macro, and more). Each market has clear resolution rules.
        </p>
        <p className={styles.p}>
          In typical <strong>binary</strong> markets (e.g. Yes / No), the traded price is often read as an{' '}
          <strong>implied probability</strong>. If “Yes” trades near $0.60, the crowd is roughly pricing a 60% chance
          that Yes wins at resolution. At resolution, winning outcome shares tend to settle toward $1 USDC per share
          and losing sides toward $0 — mechanics follow Polymarket’s rules for that market.
        </p>
        <p className={styles.p}>
          Collateral is <strong>USDC on Polygon</strong> (this app defaults to chain id 137). You need a funded wallet
          and must comply with Polymarket’s eligibility and terms.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.h2}>Gamma vs CLOB</h2>
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>API</th>
                <th>Role in this bot</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="mono">Gamma</td>
                <td>Market discovery — questions, volume, slugs. Public reads; no API key required for listing markets.</td>
              </tr>
              <tr>
                <td className="mono">CLOB</td>
                <td>
                  Order books (bids/asks), prices, placing and canceling orders. Trading uses HMAC auth and keys from
                  the Builder Program.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section className={styles.section}>
        <h2 className={styles.h2}>Using this app</h2>
        <ol className={styles.list}>
          <li>
            Prepare a <strong>JSON</strong> config (same keys as <span className="mono">config.example.yaml</span>; convert
            from YAML if needed). Use Settings → Browse to load it.
          </li>
          <li>
            Open <strong>Settings</strong>, browse to your config, set <strong>CLOB</strong> and <strong>Gamma</strong>{' '}
            hosts (defaults match Polymarket production).
          </li>
          <li>
            Add <strong>API key, secret, passphrase</strong> from{' '}
            <a className={styles.link} href="https://builders.polymarket.com" target="_blank" rel="noreferrer">
              builders.polymarket.com
            </a>
            . Optionally set <span className="mono">POLYMARKET_*</span> env vars instead of storing secrets in the file.
          </li>
          <li>
            Choose a <strong>strategy</strong> and <strong>risk</strong> limits. Enable <strong>dry run</strong> until
            you are satisfied with behavior.
          </li>
          <li>
            Press <strong>Start Bot</strong> on the Dashboard header. Watch <strong>Markets</strong>,{' '}
            <strong>Dashboard</strong> signals, and <strong>Positions</strong> as it runs.
          </li>
        </ol>
        <p className={styles.note}>
          Prefer <strong>dry run</strong> first: it exercises discovery and order-book reads without live order
          submission (depending on implementation, orders may be logged only).
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.h2}>Strategies</h2>
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Name</th>
                <th>Behavior</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>monitor</td>
                <td>Read-only: polls markets and books, no trading. Best for connectivity checks.</td>
              </tr>
              <tr>
                <td>spread</td>
                <td>
                  Posts limit bids and asks around mid-price. Tune <span className="mono">spread_bps</span> and{' '}
                  <span className="mono">order_size_usdc</span>. Inventory and adverse selection are real risks.
                </td>
              </tr>
              <tr>
                <td>momentum</td>
                <td>
                  Trades on short-term price moves vs a lookback. Tune{' '}
                  <span className="mono">momentum_threshold</span>, <span className="mono">momentum_lookback</span>, and
                  size.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p className={styles.p}>
          Use <span className="mono">market_keywords</span> to only track questions containing those words; leave empty to
          follow top markets by volume up to <span className="mono">max_markets</span>.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.h2}>Risk parameters</h2>
        <ul className={styles.list}>
          <li>
            <span className="mono">max_exposure_usdc</span> — total exposure cap across markets.
          </li>
          <li>
            <span className="mono">max_market_exposure_usdc</span> — per-market cap.
          </li>
          <li>
            <span className="mono">stop_loss_usdc</span> — halt when realized P&amp;L falls below this (negative =
            max loss).
          </li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2 className={styles.h2}>Trading vocabulary</h2>
        <ul className={styles.list}>
          <li>
            <strong>Bid / ask</strong> — best prices to sell vs buy; <strong>spread</strong> is the gap between them.
          </li>
          <li>
            <strong>Mid-price</strong> — midpoint of best bid and ask; a simple fair-value estimate.
          </li>
          <li>
            <strong>Limit order</strong> — rests in the book at your price; may not fill if price moves away.
          </li>
          <li>
            <strong>Liquidity</strong> — size available near top of book; low liquidity means more slippage for your
            size.
          </li>
          <li>
            <strong>Resolution</strong> — when the outcome is determined; payouts apply to winning shares.
          </li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2 className={styles.h2}>Disclaimer</h2>
        <p className={styles.p}>
          This tool is for education. Markets involve risk of loss. Bugs, API changes, and model risk can cause
          unexpected behavior. Only trade what you can afford to lose.
        </p>
      </section>
    </div>
  )
}
