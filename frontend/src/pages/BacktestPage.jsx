import React, { useState } from 'react'
import {
  Chart as ChartJS, CategoryScale, LinearScale,
  PointElement, LineElement, Filler, Tooltip
} from 'chart.js'
import { Line } from 'react-chartjs-2'
import Card from '../components/Card'
import { runBacktest } from '../lib/wails'
import { fmtPnl, fmtPct } from '../lib/format'
import styles from './BacktestPage.module.css'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip)

const DEFAULTS = {
  strategy: 'momentum', ticks: 500, vol: 0.015,
  spreadBps: 50, orderSize: 10, slippage: 0.001, feeBps: 100, compare: false,
}

export default function BacktestPage() {
  const [form, setForm]       = useState(DEFAULTS)
  const [result, setResult]   = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError]     = useState(null)

  const set = (k, v) => setForm(f => ({ ...f, [k]: v }))

  const run = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await runBacktest(form)
      setResult(res)
    } catch (e) {
      setError(e?.message ?? String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.grid}>
        <BacktestForm form={form} set={set} onRun={run} loading={loading} />
        {result && <BacktestResults result={result} />}
        {error  && <div className={styles.error}>{error}</div>}
      </div>
    </div>
  )
}

function BacktestForm({ form, set, onRun, loading }) {
  return (
    <Card title="Backtest Configuration">
      <div className={styles.form}>
        <Field label="Strategy">
          <select className={styles.input} value={form.strategy} onChange={e => set('strategy', e.target.value)}>
            <option value="monitor">Monitor</option>
            <option value="spread">Spread</option>
            <option value="momentum">Momentum</option>
          </select>
        </Field>
        <Field label="Ticks">
          <input className={styles.input} type="number" value={form.ticks}
            onChange={e => set('ticks', +e.target.value)} min={50} max={5000} />
        </Field>
        <Field label="Volatility">
          <input className={styles.input} type="number" step="0.001" value={form.vol}
            onChange={e => set('vol', +e.target.value)} />
        </Field>
        <Field label="Order Size (USDC)">
          <input className={styles.input} type="number" value={form.orderSize}
            onChange={e => set('orderSize', +e.target.value)} />
        </Field>
        <Field label="Spread (bps)">
          <input className={styles.input} type="number" value={form.spreadBps}
            onChange={e => set('spreadBps', +e.target.value)} />
        </Field>
        <Field label="Fee (bps)">
          <input className={styles.input} type="number" value={form.feeBps}
            onChange={e => set('feeBps', +e.target.value)} />
        </Field>
        <Field label="Slippage">
          <input className={styles.input} type="number" step="0.001" value={form.slippage}
            onChange={e => set('slippage', +e.target.value)} />
        </Field>
        <Field label="Compare all strategies">
          <input type="checkbox" checked={form.compare}
            onChange={e => set('compare', e.target.checked)} />
        </Field>
        <button className={styles.runBtn} onClick={onRun} disabled={loading}>
          {loading ? '⟳ Running…' : '▶ Run Backtest'}
        </button>
      </div>
    </Card>
  )
}

function BacktestResults({ result }) {
  if (result.comparisons?.length) {
    return <CompareResults rows={result.comparisons} />
  }

  const labels = result.equityCurve?.map((_, i) => i) ?? []
  const pnlPositive = (result.pnl ?? 0) >= 0

  const chartData = {
    labels,
    datasets: [{
      data: result.equityCurve ?? [],
      borderColor: pnlPositive ? '#30d98b' : '#f05464',
      backgroundColor: pnlPositive ? 'rgba(48,217,139,0.08)' : 'rgba(240,84,100,0.08)',
      borderWidth: 1.5,
      pointRadius: 0,
      fill: true,
      tension: 0.3,
    }],
  }

  const chartOpts = {
    responsive: true, maintainAspectRatio: false, animation: false,
    plugins: { legend: { display: false }, tooltip: { mode: 'index', intersect: false } },
    scales: {
      x: { display: false },
      y: { grid: { color: '#1e2235' }, ticks: { color: '#8b92a8', font: { size: 10 } } },
    },
  }

  return (
    <div className={styles.results}>
      <Card title={`Results — ${result.strategy}`}>
        <div className={styles.kpis}>
          <Kpi label="P&L"        value={fmtPnl(result.pnl)}      color={pnlPositive ? 'green' : 'red'} />
          <Kpi label="Trades"     value={result.trades} />
          <Kpi label="Win Rate"   value={fmtPct(result.winRate)} />
          <Kpi label="Max DD"     value={fmtPnl(result.maxDrawdown)} color="red" />
        </div>
        <div className={styles.chart}>
          <Line data={chartData} options={chartOpts} />
        </div>
      </Card>
    </div>
  )
}

function CompareResults({ rows }) {
  return (
    <Card title="Strategy Comparison">
      <table className={styles.cmpTable}>
        <thead>
          <tr>
            <th>Strategy</th><th>P&L</th><th>Trades</th><th>Win Rate</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr key={i}>
              <td>{['🥇','🥈','🥉'][i] ?? '  '} {r.name}</td>
              <td className={r.pnl >= 0 ? 'text-green' : 'text-red'}>{fmtPnl(r.pnl)}</td>
              <td>{r.trades}</td>
              <td>{fmtPct(r.winRate)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  )
}

function Field({ label, children }) {
  return (
    <label className={styles.field}>
      <span className={styles.fieldLabel}>{label}</span>
      {children}
    </label>
  )
}

function Kpi({ label, value, color }) {
  return (
    <div className={styles.kpi}>
      <div className={styles.kpiLabel}>{label}</div>
      <div className={`${styles.kpiValue} ${color ? styles[color] : ''}`}>{value}</div>
    </div>
  )
}
