import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { fetchPlayerProfile } from '../api/players'
import { FaceitStatsPanel } from '../components/FaceitStatsPanel'
import { MapCanvas } from '../components/MapCanvas'
import { MetricsDashboard } from '../components/MetricsDashboard'
import type { PlayerProfile } from '../types'

export function PlayerProfilePage() {
  const { playerId } = useParams<{ playerId: string }>()
  const [profile, setProfile] = useState<PlayerProfile | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!playerId) {
      return
    }
    let cancelled = false

    fetchPlayerProfile(playerId)
      .then((data) => {
        if (!cancelled) {
          setProfile(data)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'failed to load player profile')
        }
      })

    return () => {
      cancelled = true
    }
  }, [playerId])

  if (!playerId) {
    return <div className="p-8 text-sm text-red-600">Missing player id</div>
  }

  if (error) {
    return <div className="p-8 text-sm text-red-600">{error}</div>
  }

  if (!profile) {
    return <div className="p-8 text-sm text-slate-400">Loading player profile</div>
  }

  return (
    <div className="min-h-screen bg-slate-50 p-8">
      <div className="mx-auto max-w-4xl space-y-6">
        <header>
          <h1 className="text-xl font-semibold text-slate-900">{profile.playerName}</h1>
          <p className="text-sm text-slate-500">{profile.map.name || 'Unknown map'}</p>
        </header>

        <MetricsDashboard metrics={profile.metrics} />

        <MapCanvas
          mapImageUrl={profile.map.imageUrl || '/placeholder-map.svg'}
          bounds={profile.map.bounds}
          points={profile.positions}
        />

        <FaceitStatsPanel playerId={profile.playerId} />
      </div>
    </div>
  )
}
