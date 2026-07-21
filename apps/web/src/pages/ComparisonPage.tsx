import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { fetchPlayerProfile } from '../api/players'
import { RadarChart } from '../components/RadarChart'
import type { RadarPlayerData } from '../types'

export function ComparisonPage() {
  const { playerAId, playerBId } = useParams<{ playerAId: string; playerBId: string }>()
  const [playerA, setPlayerA] = useState<RadarPlayerData | null>(null)
  const [playerB, setPlayerB] = useState<RadarPlayerData | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!playerAId || !playerBId) {
      return
    }
    let cancelled = false

    Promise.all([fetchPlayerProfile(playerAId), fetchPlayerProfile(playerBId)])
      .then(([profileA, profileB]) => {
        if (cancelled) {
          return
        }
        setPlayerA({ playerId: profileA.playerId, playerName: profileA.playerName, metrics: profileA.metrics })
        setPlayerB({ playerId: profileB.playerId, playerName: profileB.playerName, metrics: profileB.metrics })
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'failed to load comparison')
        }
      })

    return () => {
      cancelled = true
    }
  }, [playerAId, playerBId])

  if (!playerAId || !playerBId) {
    return <div className="p-8 text-sm text-red-600">Missing player ids</div>
  }

  if (error) {
    return <div className="p-8 text-sm text-red-600">{error}</div>
  }

  if (!playerA || !playerB) {
    return <div className="p-8 text-sm text-slate-400">Loading comparison</div>
  }

  return (
    <div className="min-h-screen bg-slate-50 p-8">
      <div className="mx-auto max-w-2xl space-y-6">
        <header>
          <h1 className="text-xl font-semibold text-slate-900">
            {playerA.playerName} vs {playerB.playerName}
          </h1>
        </header>
        <RadarChart playerA={playerA} playerB={playerB} />
      </div>
    </div>
  )
}
