import React, { useState, useMemo } from 'react'
import Card from '../components/Card'
import MarketsTable from '../components/MarketsTable'
import styles from './MarketsPage.module.css'

export default function MarketsPage({ markets }) {
  const [query, setQuery]   = useState('')
  const [filter, setFilter] = useState('all')

  const filtered = useMemo(() => {
    return (markets ?? []).filter(m => {
      const matchText = m.question.toLowerCase().includes(query.toLowerCase())
      const matchSig  = filter === 'all'
        || (filter === 'active' && m.signal !== 'NONE' && m.signal !== '')
      return matchText && matchSig
    })
  }, [markets, query, filter])

  const actionBar = (
    <div className={styles.controls}>
      <input
        className={styles.search}
        placeholder="Search markets…"
        value={query}
        onChange={e => setQuery(e.target.value)}
      />
      <div className={styles.tabs}>
        {['all', 'active'].map(f => (
          <button
            key={f}
            className={`${styles.tab} ${filter === f ? styles.tabActive : ''}`}
            onClick={() => setFilter(f)}
          >
            {f === 'all' ? 'All' : 'With Signal'}
          </button>
        ))}
      </div>
    </div>
  )

  return (
    <div className={styles.page}>
      <Card title={`Markets (${filtered.length})`} action={actionBar}>
        <MarketsTable markets={filtered} />
      </Card>
    </div>
  )
}
