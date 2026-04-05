import React from 'react'
import styles from './StatCard.module.css'

/**
 * A single KPI tile shown in the top row of the dashboard.
 * @param {string}  label   - metric name
 * @param {string}  value   - primary value string
 * @param {string}  [sub]   - secondary line
 * @param {string}  [color] - 'green' | 'red' | 'yellow' | 'accent'
 * @param {string}  [icon]  - emoji / unicode glyph
 */
export default function StatCard({ label, value, sub, color, icon }) {
  return (
    <div className={styles.card}>
      <div className={styles.top}>
        <span className={styles.label}>{label}</span>
        {icon && <span className={styles.icon}>{icon}</span>}
      </div>
      <div className={`${styles.value} ${color ? styles[color] : ''}`}>{value}</div>
      {sub && <div className={styles.sub}>{sub}</div>}
    </div>
  )
}
