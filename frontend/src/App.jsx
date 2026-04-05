import React, { useState } from 'react'
import Layout from './components/Layout'
import Header from './components/Header'
import DashboardPage  from './pages/DashboardPage'
import MarketsPage    from './pages/MarketsPage'
import PositionsPage  from './pages/PositionsPage'
import BacktestPage   from './pages/BacktestPage'
import SettingsPage   from './pages/SettingsPage'
import DocsPage       from './pages/DocsPage'
import { useBotState } from './hooks/useBotState'
import { useConfig }   from './hooks/useConfig'

const PAGE_TITLES = {
  dashboard: 'Dashboard',
  markets:   'Live Markets',
  positions: 'Positions',
  backtest:  'Backtest',
  settings:  'Settings',
  docs:      'Documentation',
}

export default function App() {
  const [page, setPage] = useState('dashboard')
  const { state, running, loading, start, stop } = useBotState()
  const { cfg, configPath, loadErr, pickAndLoad, update, updateStrategy, updateRisk } = useConfig()

  const handleStart = () => start(cfg)

  const handleSettingsStart = () => {
    setPage('dashboard')
    start(cfg)
  }

  return (
    <Layout
      page={page}
      onNav={setPage}
      running={running}
      dryRun={state.dryRun || cfg.dry_run}
    >
      <Header
        title={PAGE_TITLES[page]}
        running={running}
        loading={loading}
        dryRun={state.dryRun || cfg.dry_run}
        uptime={state.uptime}
        lastUpdate={state.lastUpdate}
        onStart={handleStart}
        onStop={stop}
      />

      {page === 'dashboard' && (
        <DashboardPage
          state={state}
          maxExposure={cfg.risk?.max_exposure_usdc ?? 500}
        />
      )}

      {page === 'markets' && (
        <MarketsPage markets={state.markets} />
      )}

      {page === 'positions' && (
        <PositionsPage positions={state.positions} />
      )}

      {page === 'backtest' && (
        <BacktestPage />
      )}

      {page === 'settings' && (
        <SettingsPage
          cfg={cfg}
          configPath={configPath}
          onPick={pickAndLoad}
          onUpdate={update}
          onUpdateStrategy={updateStrategy}
          onUpdateRisk={updateRisk}
          onStart={handleSettingsStart}
          running={running}
        />
      )}

      {page === 'docs' && <DocsPage />}
    </Layout>
  )
}
