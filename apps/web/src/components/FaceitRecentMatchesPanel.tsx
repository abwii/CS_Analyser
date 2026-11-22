import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { fetchFaceitRecentMatches } from '../api/players'
import type { FaceitRecentMatches } from '../types'

interface FaceitRecentMatchesPanelProps {
  playerId: string
}

export function FaceitRecentMatchesPanel({ playerId }: FaceitRecentMatchesPanelProps) {
  const [matches, setMatches] = useState<FaceitRecentMatches | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    fetchFaceitRecentMatches(playerId)
      .then((data) => {
        if (!cancelled) {
          setMatches(data)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'failed to load faceit matches')
        }
      })

    return () => {
      cancelled = true
    }
  }, [playerId])

  return (
    <div className="rounded-lg border border-slate-200 bg-white">
      <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Faceit recent matches</h2>
        {matches?.available && matches.faceitNickname && (
          <span className="text-sm text-slate-500">{matches.faceitNickname}</span>
        )}
      </div>

      {error && <p className="px-4 py-3 text-sm text-red-600">{error}</p>}

      {!error && !matches && <p className="px-4 py-3 text-sm text-slate-400">Loading faceit matches</p>}

      {!error && matches && !matches.available && (
        <p className="px-4 py-3 text-sm text-slate-400">No linked faceit account found for this player</p>
      )}

      {!error && matches?.available && matches.matches.length === 0 && (
        <p className="px-4 py-3 text-sm text-slate-400">No recent matches with a known map</p>
      )}

      {!error && matches?.available && matches.matches.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-slate-200 text-slate-500">
              <tr>
                <th className="px-4 py-2 font-medium">Map</th>
                <th className="px-4 py-2 font-medium">Result</th>
                <th className="px-4 py-2 font-medium" />
              </tr>
            </thead>
            <tbody>
              {matches.matches.map((match) => (
                <tr key={match.matchId} className="border-b border-slate-100 last:border-0">
                  <td className="px-4 py-2 text-slate-800">{match.map}</td>
                  <td className={`px-4 py-2 ${match.won ? 'text-emerald-600' : 'text-red-600'}`}>
                    {match.won ? 'Won' : 'Lost'}
                  </td>
                  <td className="px-4 py-2 text-right">
                    <Link
                      to="/demos"
                      title="Faceit demo downloads require access this project doesn't have yet - upload the .dem manually"
                      className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700 hover:bg-slate-50"
                    >
                      Analyser
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
