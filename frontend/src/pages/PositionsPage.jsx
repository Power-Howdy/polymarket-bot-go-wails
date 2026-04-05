import React from 'react'
import Card from '../components/Card'
import PositionsTable from '../components/PositionsTable'
import StatCard from '../components/StatCard'
import { fmtPnl, pnlClass } from '../lib/format'
import styles from './PositionsPage.module.css'

export default function PositionsPage({ positions }) {
  const totalRealized   = (positions ?? []).reduce((s, p) => s + p.realizedPnL, 0)
  const totalUnrealized = (positions ?? []).reduce((s, p) => s + p.unrealized, 0)
  const totalPnL        = totalRealized + totalUnrealized

  return (
    <div className={styles.page}>
      <div className={styles.stats}>
        <StatCard
          label="Realized P&L"
          value={fmtPnl(totalRealized)}
          color={totalRealized >= 0 ? 'green' : 'red'}
        />
        <StatCard
          label="Unrealized P&L"
          value={fmtPnl(totalUnrealized)}
          color={totalUnrealized >= 0 ? 'green' : 'red'}
        />
        <StatCard
          label="Total P&L"
          value={fmtPnl(totalPnL)}
          color={totalPnL >= 0 ? 'green' : 'red'}
        />
        <StatCard
          label="Open Positions"
          value={String(positions?.length ?? 0)}
        />
      </div>

      <Card title="Open Positions">
        <PositionsTable positions={positions} />
      </Card>
    </div>
  )
}
