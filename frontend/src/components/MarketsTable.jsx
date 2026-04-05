import React from 'react'
import { trunc, fmtUsd, signalBadge } from '../lib/format'
import styles from './MarketsTable.module.css'

export default function MarketsTable({ markets }) {
  if (!markets?.length) return <Empty />

  return (
    <div className={styles.wrap}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th>Market</th>
            <th className={styles.num}>Bid</th>
            <th className={styles.num}>Ask</th>
            <th className={styles.num}>Mid</th>
            <th className={styles.num}>Spread</th>
            <th className={styles.num}>Volume</th>
            <th>Signal</th>
          </tr>
        </thead>
        <tbody>
          {markets.map((m, i) => <MarketRow key={i} m={m} />)}
        </tbody>
      </table>
    </div>
  )
}

function MarketRow({ m }) {
  const side = m.signal?.startsWith('BUY') ? 'BUY' : m.signal?.startsWith('SELL') ? 'SELL' : 'NONE'
  return (
    <tr className={styles.row}>
      <td className={styles.question} title={m.question}>
        {trunc(m.question, 52)}
        <span className={styles.outcome}>{m.outcome}</span>
      </td>
      <td className={`${styles.num} mono`}>{m.bid.toFixed(3)}</td>
      <td className={`${styles.num} mono`}>{m.ask.toFixed(3)}</td>
      <td className={`${styles.num} mono`}>{m.mid.toFixed(3)}</td>
      <td className={`${styles.num} mono text-dim`}>{m.spread.toFixed(4)}</td>
      <td className={`${styles.num} mono text-secondary`}>{fmtUsd(m.volume)}</td>
      <td>
        <span className={`badge ${signalBadge(side)}`}>{m.signal || 'NONE'}</span>
      </td>
    </tr>
  )
}

function Empty() {
  return <div className={styles.empty}>Waiting for market data…</div>
}
