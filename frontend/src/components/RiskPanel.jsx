import React from 'react'
import { fmtPnl, fmtUsd, pnlClass } from '../lib/format'
import styles from './RiskPanel.module.css'

export default function RiskPanel({ risk, maxExposure = 500 }) {
  const r = risk ?? {}
  const exposurePct = Math.min(100, ((r.totalExposure ?? 0) / maxExposure) * 100)
  const barColor = exposurePct > 80 ? 'red' : exposurePct > 50 ? 'yellow' : 'green'

  return (
    <div className={styles.panel}>
      <Row label="Total Exposure">
        <span className="mono">{fmtUsd(r.totalExposure ?? 0)}</span>
        <span className={styles.max}>/ {fmtUsd(maxExposure)}</span>
      </Row>

      <div className={styles.barTrack}>
        <div
          className={`${styles.barFill} ${styles[barColor]}`}
          style={{ width: `${exposurePct}%` }}
        />
      </div>

      <Row label="Realized P&L">
        <span className={`mono ${pnlClass(r.realizedPnL ?? 0)}`}>
          {fmtPnl(r.realizedPnL ?? 0)}
        </span>
      </Row>

      <div className={styles.divider} />

      <Row label="Total Trades">
        <span className="mono">{r.tradesTotal ?? 0}</span>
      </Row>
      <Row label="Buys / Sells">
        <span className="mono text-green">{r.tradesBuys ?? 0}</span>
        <span className={styles.slash}>/</span>
        <span className="mono text-red">{r.tradesSells ?? 0}</span>
      </Row>
    </div>
  )
}

function Row({ label, children }) {
  return (
    <div className={styles.row}>
      <span className={styles.label}>{label}</span>
      <div className={styles.value}>{children}</div>
    </div>
  )
}
