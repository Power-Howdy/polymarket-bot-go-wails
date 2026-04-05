import React from 'react'
import { trunc, signalBadge } from '../lib/format'
import styles from './SignalsFeed.module.css'

export default function SignalsFeed({ signals }) {
  if (!signals?.length) return <div className={styles.empty}>No signals yet.</div>

  return (
    <ul className={styles.list}>
      {signals.map((s, i) => <SignalRow key={i} s={s} />)}
    </ul>
  )
}

function SignalRow({ s }) {
  return (
    <li className={styles.row}>
      <span className={`badge ${signalBadge(s.side)}`}>{s.side}</span>
      <div className={styles.body}>
        <div className={styles.market}>{trunc(s.market, 56)}</div>
        <div className={styles.reason}>{s.reason}</div>
      </div>
      {s.side !== 'NONE' && (
        <div className={styles.price}>
          <span className={styles.priceVal}>${s.price.toFixed(3)}</span>
          <span className={styles.priceLabel}>× {s.size.toFixed(1)}</span>
        </div>
      )}
    </li>
  )
}
