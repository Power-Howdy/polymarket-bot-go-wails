import React from 'react'
import Card from '../components/Card'
import styles from './SettingsPage.module.css'

export default function SettingsPage({ cfg, configPath, onPick, onUpdate, onUpdateStrategy, onUpdateRisk, onStart, running }) {
  return (
    <div className={styles.page}>
      <div className={styles.grid}>
        <ConnectionCard cfg={cfg} configPath={configPath} onPick={onPick} onUpdate={onUpdate} />
        <StrategyCard cfg={cfg.strategy} onChange={onUpdateStrategy} />
        <RiskCard cfg={cfg.risk} onChange={onUpdateRisk} />
        <StartCard running={running} dryRun={cfg.dry_run} onUpdate={onUpdate} onStart={onStart} />
      </div>
    </div>
  )
}

function ConnectionCard({ cfg, configPath, onPick, onUpdate }) {
  return (
    <Card title="Connection">
      <div className={styles.fields}>
        <Field label="Config file">
          <div className={styles.fileRow}>
            <span className={styles.filePath}>{configPath || 'No file loaded'}</span>
            <button className={styles.btn} onClick={onPick}>Browse…</button>
          </div>
        </Field>
        <Field label="CLOB Host">
          <input className={styles.input} value={cfg.clob_host ?? ''}
            onChange={e => onUpdate({ clob_host: e.target.value })} />
        </Field>
        <Field label="Gamma Host">
          <input className={styles.input} value={cfg.gamma_host ?? ''}
            onChange={e => onUpdate({ gamma_host: e.target.value })} />
        </Field>
        <Field label="API Key">
          <input className={styles.input} type="password" value={cfg.api_key ?? ''}
            onChange={e => onUpdate({ api_key: e.target.value })} placeholder="••••••••" />
        </Field>
        <Field label="API Secret">
          <input className={styles.input} type="password" value={cfg.api_secret ?? ''}
            onChange={e => onUpdate({ api_secret: e.target.value })} placeholder="••••••••" />
        </Field>
        <Field label="Poll Interval">
          <input className={styles.input} value={cfg.poll_interval ?? '10s'}
            onChange={e => onUpdate({ poll_interval: e.target.value })} />
        </Field>
      </div>
    </Card>
  )
}

function StrategyCard({ cfg, onChange }) {
  return (
    <Card title="Strategy">
      <div className={styles.fields}>
        <Field label="Strategy">
          <select className={styles.input} value={cfg.name}
            onChange={e => onChange({ name: e.target.value })}>
            <option value="monitor">Monitor (no orders)</option>
            <option value="spread">Spread Maker</option>
            <option value="momentum">Momentum</option>
          </select>
        </Field>
        <Field label="Max Markets">
          <input className={styles.input} type="number" value={cfg.max_markets}
            onChange={e => onChange({ max_markets: +e.target.value })} min={1} max={20} />
        </Field>
        <Field label="Order Size (USDC)">
          <input className={styles.input} type="number" value={cfg.order_size_usdc}
            onChange={e => onChange({ order_size_usdc: +e.target.value })} />
        </Field>
        <Field label="Spread (bps)">
          <input className={styles.input} type="number" value={cfg.spread_bps}
            onChange={e => onChange({ spread_bps: +e.target.value })} />
        </Field>
        <Field label="Momentum Threshold">
          <input className={styles.input} type="number" step="0.005" value={cfg.momentum_threshold}
            onChange={e => onChange({ momentum_threshold: +e.target.value })} />
        </Field>
        <Field label="Momentum Lookback">
          <input className={styles.input} type="number" value={cfg.momentum_lookback}
            onChange={e => onChange({ momentum_lookback: +e.target.value })} min={1} />
        </Field>
      </div>
    </Card>
  )
}

function RiskCard({ cfg, onChange }) {
  return (
    <Card title="Risk Limits">
      <div className={styles.fields}>
        <Field label="Max Total Exposure (USDC)">
          <input className={styles.input} type="number" value={cfg.max_exposure_usdc}
            onChange={e => onChange({ max_exposure_usdc: +e.target.value })} />
        </Field>
        <Field label="Max Market Exposure (USDC)">
          <input className={styles.input} type="number" value={cfg.max_market_exposure_usdc}
            onChange={e => onChange({ max_market_exposure_usdc: +e.target.value })} />
        </Field>
        <Field label="Stop Loss (USDC, negative)">
          <input className={styles.input} type="number" value={cfg.stop_loss_usdc}
            onChange={e => onChange({ stop_loss_usdc: +e.target.value })} />
        </Field>
      </div>
    </Card>
  )
}

function StartCard({ running, dryRun, onUpdate, onStart }) {
  return (
    <Card title="Run">
      <div className={styles.fields}>
        <label className={styles.checkRow}>
          <input type="checkbox" checked={dryRun}
            onChange={e => onUpdate({ dry_run: e.target.checked })} />
          <span>Dry Run (no real orders)</span>
        </label>
        <button className={styles.startBtn} onClick={onStart} disabled={running}>
          {running ? '⟳ Bot is running…' : '▶ Apply & Start Bot'}
        </button>
      </div>
    </Card>
  )
}

function Field({ label, children }) {
  return (
    <label className={styles.field}>
      <span className={styles.label}>{label}</span>
      {children}
    </label>
  )
}
