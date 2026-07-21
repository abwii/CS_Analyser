import type { PlayerMetrics } from '../types'

interface MetricsDashboardProps {
  metrics: PlayerMetrics
}

function MetricTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-4">
      <p className="text-xs font-medium uppercase tracking-wide text-slate-500">{label}</p>
      <p className="mt-1 text-2xl font-semibold text-slate-900">{value}</p>
    </div>
  )
}

export function MetricsDashboard({ metrics }: MetricsDashboardProps) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
      <MetricTile label="KPR" value={metrics.kpr.toFixed(2)} />
      <MetricTile label="DPR" value={metrics.dpr.toFixed(2)} />
      <MetricTile label="KAST" value={`${metrics.kast.toFixed(1)}%`} />
      <MetricTile label="Impact" value={metrics.impact.toFixed(2)} />
      <MetricTile label="ADR" value={metrics.adr.toFixed(1)} />
    </div>
  )
}
