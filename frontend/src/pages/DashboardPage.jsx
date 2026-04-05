import React from 'react'
import StatCard from '../components/StatCard'
import Card from '../components/Card'
import MarketsTable from '../components/MarketsTable'
import SignalsFeed from '../components/SignalsFeed'
import RiskPanel from '../components/RiskPanel'
import ErrorBanner from '../components/ErrorBanner'
import { fmtPnl, fmtUsd } from '../lib/format'
import styles from './DashboardPage.module.css'

export default function DashboardPage({ state, maxExposure }) {
  const r = state.risk

  return (
    <div className={styles.page}>
      <ErrorBanner errors={state.errors} />

      {/* ── KPI row ──────────────────────────────────────────── */}
      <div className={styles.stats}>
        <StatCard
          label="Realized P&L"
          value={fmtPnl(r.realizedPnL)}
          color={r.realizedPnL >= 0 ? 'green' : 'red'}
          icon="$"
        />
        <StatCard
          label="Exposure"
          value={fmtUsd(r.totalExposure)}
          sub={`of ${fmtUsd(maxExposure)} limit`}
          icon="⊡"
        />
        <StatCard
          label="Trades"
          value={String(r.tradesTotal)}
          sub={`${r.tradesBuys} buys · ${r.tradesSells} sells`}
          icon="⇅"
        />
        <StatCard
          label="Markets"
          value={String(state.markets?.length ?? 0)}
          sub={`strategy: ${state.strategy}`}
          icon="◈"
        />
        <StatCard
          label="Positions"
          value={String(state.positions?.length ?? 0)}
          icon="◉"
        />
      </div>

      {/* ── Main grid ────────────────────────────────────────── */}
      <div className={styles.grid}>
        <div className={styles.colWide}>
          <Card title="Live Markets">
            <MarketsTable markets={state.markets} />
          </Card>
        </div>

        <div className={styles.colNarrow}>
          <Card title="Recent Signals">
            <SignalsFeed signals={state.recentSignals} />
          </Card>
          <Card title="Risk Summary">
            <RiskPanel risk={r} maxExposure={maxExposure} />
          </Card>
        </div>
      </div>
    </div>
  )
}
