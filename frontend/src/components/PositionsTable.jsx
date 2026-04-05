import React from 'react'
import { trunc, fmtPnl, pnlClass } from '../lib/format'
import styles from './PositionsTable.module.css'

export default function PositionsTable({ positions }) {
  if (!positions?.length) return <div className={styles.empty}>No open positions.</div>

  return (
    <div className={styles.wrap}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th>Market (Outcome)</th>
            <th className={styles.num}>Size</th>
            <th className={styles.num}>Avg Cost</th>
            <th className={styles.num}>Realized P&L</th>
            <th className={styles.num}>Unrealized</th>
            <th className={styles.num}>Total P&L</th>
          </tr>
        </thead>
        <tbody>
          {positions.map((p, i) => <PositionRow key={i} p={p} />)}
        </tbody>
      </table>
    </div>
  )
}

function PositionRow({ p }) {
  const total = p.realizedPnL + p.unrealized
  return (
    <tr className={styles.row}>
      <td>
        <div className={styles.market}>{trunc(p.market, 48)}</div>
        <div className={styles.outcome}>{p.outcome}</div>
      </td>
      <td className={`${styles.num} mono`}>{p.totalSize.toFixed(2)}</td>
      <td className={`${styles.num} mono`}>${p.avgBuyPrice.toFixed(4)}</td>
      <td className={`${styles.num} mono ${pnlClass(p.realizedPnL)}`}>
        {fmtPnl(p.realizedPnL)}
      </td>
      <td className={`${styles.num} mono ${pnlClass(p.unrealized)}`}>
        {fmtPnl(p.unrealized)}
      </td>
      <td className={`${styles.num} mono ${pnlClass(total)}`}>
        <strong>{fmtPnl(total)}</strong>
      </td>
    </tr>
  )
}
