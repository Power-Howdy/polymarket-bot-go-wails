import React from 'react'
import styles from './Layout.module.css'

const NAV = [
  { id: 'dashboard', label: 'Dashboard', icon: '◈' },
  { id: 'markets',   label: 'Markets',   icon: '⊞' },
  { id: 'positions', label: 'Positions', icon: '◉' },
  { id: 'backtest',  label: 'Backtest',  icon: '⟳' },
  { id: 'settings',  label: 'Settings',  icon: '⚙' },
  { id: 'docs',      label: 'Docs',      icon: '☰' },
]

export default function Layout({ page, onNav, running, dryRun, children }) {
  return (
    <div className={styles.shell}>
      <aside className={styles.sidebar}>
        <div className={styles.logo}>
          <span className={styles.logoIcon}>◈</span>
          <span className={styles.logoText}>PolyBot</span>
        </div>

        <nav className={styles.nav}>
          {NAV.map(n => (
            <button
              key={n.id}
              className={`${styles.navBtn} ${page === n.id ? styles.active : ''}`}
              onClick={() => onNav(n.id)}
            >
              <span className={styles.navIcon}>{n.icon}</span>
              <span>{n.label}</span>
            </button>
          ))}
        </nav>

        <div className={styles.statusArea}>
          <div className={`${styles.statusDot} ${running ? styles.dotGreen : styles.dotGrey}`} />
          <span className={styles.statusLabel}>
            {running ? (dryRun ? 'Dry Run' : 'Live') : 'Stopped'}
          </span>
        </div>
      </aside>

      <main className={styles.main}>
        {children}
      </main>
    </div>
  )
}
