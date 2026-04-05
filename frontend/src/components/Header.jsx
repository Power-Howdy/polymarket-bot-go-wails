import React from 'react'
import styles from './Header.module.css'

export default function Header({ title, running, loading, dryRun, uptime, lastUpdate, onStart, onStop }) {
  return (
    <header className={styles.header}>
      <div className={styles.left}>
        <h1 className={styles.title}>{title}</h1>
        {dryRun && <span className="badge badge-yellow">DRY RUN</span>}
      </div>

      <div className={styles.meta}>
        {running && (
          <>
            <Stat label="Uptime"   value={uptime} />
            <Stat label="Updated"  value={lastUpdate} />
          </>
        )}
      </div>

      <div className={styles.right}>
        {running ? (
          <button className={styles.btnStop} onClick={onStop} disabled={loading}>
            ■ Stop Bot
          </button>
        ) : (
          <button className={styles.btnStart} onClick={onStart} disabled={loading}>
            {loading ? '…' : '▶ Start Bot'}
          </button>
        )}
      </div>
    </header>
  )
}

function Stat({ label, value }) {
  return (
    <div className={styles.stat}>
      <span className={styles.statLabel}>{label}</span>
      <span className={styles.statValue}>{value}</span>
    </div>
  )
}
