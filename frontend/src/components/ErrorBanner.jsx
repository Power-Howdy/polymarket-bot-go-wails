import React, { useState, useEffect } from 'react'
import styles from './ErrorBanner.module.css'

export default function ErrorBanner({ errors }) {
  const [dismissed, setDismissed] = useState(new Set())

  // Reset dismissed list when new errors arrive
  useEffect(() => { setDismissed(new Set()) }, [errors?.length])

  const visible = (errors ?? []).filter((_, i) => !dismissed.has(i))
  if (!visible.length) return null

  return (
    <div className={styles.container}>
      {visible.map((msg, i) => (
        <div key={i} className={styles.banner}>
          <span className={styles.icon}>⚠</span>
          <span className={styles.msg}>{msg}</span>
          <button className={styles.close} onClick={() => setDismissed(d => new Set([...d, i]))}>
            ✕
          </button>
        </div>
      ))}
    </div>
  )
}
