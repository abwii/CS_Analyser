import { useEffect, useState } from 'react'
import { fetchFaceitStats } from '../api/players'
import type { FaceitStats } from '../types'

interface FaceitStatsPanelProps {
  playerId: string
}

export function FaceitStatsPanel({ playerId }: FaceitStatsPanelProps) {
  const [stats, setStats] = useState<FaceitStats | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    fetchFaceitStats(playerId)
      .then((data) => {
        if (!cancelled) {
          setStats(data)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'failed to load faceit stats')
        }
      })

    return () => {
      cancelled = true
    }
  }, [playerId])

  return (
    <div className="rounded-lg border border-slate-200 bg-white">
      <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Faceit by map</h2>
        {stats?.available && stats.faceitNickname && (
          <span className="text-sm text-slate-500">{stats.faceitNickname}</span>
        )}
      </div>

      {error && <p className="px-4 py-3 text-sm text-red-600">{error}</p>}

      {!error && !stats && <p className="px-4 py-3 text-sm text-slate-400">Loading faceit stats</p>}

      {!error && stats && !stats.available && (
        <p className="px-4 py-3 text-sm text-slate-400">No linked faceit account found for this player</p>
      )}

      {!error && stats?.available && stats.byMap.length === 0 && (
        <p className="px-4 py-3 text-sm text-slate-400">No recent matches with a known map</p>
      )}

      {!error && stats?.available && stats.byMap.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-slate-200 text-slate-500">
              <tr>
                <th className="px-4 py-2 font-medium">Map</th>
                <th className="px-4 py-2 font-medium">Matches</th>
                <th className="px-4 py-2 font-medium">Win rate</th>
                <th className="px-4 py-2 font-medium">Avg K</th>
                <th className="px-4 py-2 font-medium">Avg D</th>
              </tr>
            </thead>
            <tbody>
              {stats.byMap.map((mapStats) => (
                <tr key={mapStats.map} className="border-b border-slate-100 last:border-0">
                  <td className="px-4 py-2 text-slate-800">{mapStats.map}</td>
                  <td className="px-4 py-2 text-slate-600">{mapStats.matchesPlayed}</td>
                  <td className="px-4 py-2 text-slate-600">{mapStats.winRate.toFixed(0)}%</td>
                  <td className="px-4 py-2 text-slate-600">{mapStats.avgKills.toFixed(1)}</td>
                  <td className="px-4 py-2 text-slate-600">{mapStats.avgDeaths.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
