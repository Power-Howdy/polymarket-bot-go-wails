/** Format a dollar value with sign and 2 decimal places. */
export function fmtPnl(v) {
  const sign = v >= 0 ? '+' : ''
  return `${sign}$${v.toFixed(2)}`
}

/** Format a price as a percentage string (0.43 → "43.0%"). */
export function fmtPct(v) {
  return `${(v * 100).toFixed(1)}%`
}

/** Format a dollar amount. */
export function fmtUsd(v) {
  return `$${v.toFixed(2)}`
}

/** Truncate a string to n characters with an ellipsis. */
export function trunc(s, n) {
  if (!s) return ''
  return s.length <= n ? s : s.slice(0, n - 1) + '…'
}

/** Return CSS colour class based on PnL sign. */
export function pnlClass(v) {
  if (v > 0) return 'text-green'
  if (v < 0) return 'text-red'
  return 'text-secondary'
}

/** Return badge variant for a signal side string. */
export function signalBadge(side) {
  if (side === 'BUY')  return 'badge-green'
  if (side === 'SELL') return 'badge-red'
  return 'badge-dim'
}
