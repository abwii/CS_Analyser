import { useEffect, useRef } from 'react'
import { useVetoSocket } from '../hooks/useVetoSocket'
import type { VetoState } from '../types'

interface VetoBoardProps {
  wsUrl: string
  localTeam: string | null
}

function statusText(state: VetoState, localTeam: string | null, connected: boolean): string {
  if (!connected) {
    return 'Disconnected'
  }
  if (state.completed) {
    return state.deciderMap ? `Veto complete. Decider: ${state.deciderMap}` : 'Veto complete'
  }
  const turnLabel = state.currentTurn === localTeam ? 'Your turn' : `${state.currentTurn} turn`
  return `${turnLabel} to ${state.nextAction ?? 'act'}`
}

export function VetoBoard({ wsUrl, localTeam }: VetoBoardProps) {
  const { state, error, connected, sendAction } = useVetoSocket(wsUrl)
  const logEndRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [state?.history.length])

  if (!state) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50 text-sm text-slate-500">
        Connecting
      </div>
    )
  }

  const isLocalTurn = localTeam !== null && state.currentTurn === localTeam
  const locked = state.completed || !connected || !isLocalTurn

  function handleMapClick(map: string) {
    if (locked || !state?.nextAction || !localTeam) {
      return
    }
    sendAction({ team: localTeam, action: state.nextAction, map })
  }

  return (
    <div className="min-h-screen bg-slate-50 p-8">
      <div className="mx-auto max-w-3xl space-y-6">
        <header className="space-y-1">
          <h1 className="text-xl font-semibold text-slate-900">
            {state.teamA} vs {state.teamB}
          </h1>
          <p className="text-sm text-slate-500">{state.format} map veto</p>
        </header>

        <div className="rounded-lg border border-slate-200 bg-white p-4 text-sm text-slate-700">
          {statusText(state, localTeam, connected)}
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}

        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
          {state.availableMaps.map((map) => (
            <button
              key={map}
              type="button"
              disabled={locked}
              onClick={() => handleMapClick(map)}
              className="rounded-lg border border-slate-200 bg-white px-4 py-6 text-sm font-medium text-slate-800 transition enabled:hover:border-slate-400 disabled:cursor-not-allowed disabled:opacity-40"
            >
              {map}
            </button>
          ))}
        </div>

        <div className="rounded-lg border border-slate-200 bg-white">
          <h2 className="border-b border-slate-200 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-slate-500">
            History
          </h2>
          <div className="max-h-56 space-y-1 overflow-y-auto px-4 py-3 font-mono text-sm text-slate-700">
            {state.history.length === 0 && <p className="text-slate-400">No actions yet</p>}
            {state.history.map((entry, index) => (
              <p key={index}>
                {entry.team} {entry.action} {entry.map}
              </p>
            ))}
            {state.completed && state.deciderMap && <p className="text-emerald-600">decider {state.deciderMap}</p>}
            <div ref={logEndRef} />
          </div>
        </div>
      </div>
    </div>
  )
}
