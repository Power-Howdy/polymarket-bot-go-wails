import React from 'react'
import styles from './Card.module.css'

export default function Card({ title, children, action }) {
  return (
    <div className={styles.card}>
      {title && (
        <div className={styles.header}>
          <span className={styles.title}>{title}</span>
          {action && <div className={styles.action}>{action}</div>}
        </div>
      )}
      <div className={styles.body}>{children}</div>
    </div>
  )
}
