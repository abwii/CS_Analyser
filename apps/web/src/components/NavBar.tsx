import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { clearToken, getToken, isTokenValid } from '../lib/auth'

function extractSteamId64(rawInput: string): string | null {
  const input = rawInput.trim()
  if (/^\d+$/.test(input)) {
    return input
  }
  const profileUrlMatch = input.match(/steamcommunity\.com\/profiles\/(\d+)/)
  if (profileUrlMatch) {
    return profileUrlMatch[1]
  }
  return null
}

export function NavBar() {
  const navigate = useNavigate()
  const [playerId, setPlayerId] = useState('')
  const [compareA, setCompareA] = useState('')
  const [compareB, setCompareB] = useState('')
  const [sessionId, setSessionId] = useState('')
  const [team, setTeam] = useState('')
  const [error, setError] = useState<string | null>(null)

  const loggedIn = isTokenValid(getToken())

  function handleLogout() {
    clearToken()
    navigate('/login', { replace: true })
  }

  function goToPlayer(event: React.FormEvent) {
    event.preventDefault()
    const id = extractSteamId64(playerId)
    if (!id) {
      setError('Profile: enter a numeric SteamID64, not a profile URL or nickname')
      return
    }
    setError(null)
    navigate(`/players/${id}`)
  }

  function goToCompare(event: React.FormEvent) {
    event.preventDefault()
    const idA = extractSteamId64(compareA)
    const idB = extractSteamId64(compareB)
    if (!idA || !idB) {
      setError('Compare: both fields need a numeric SteamID64')
      return
    }
    setError(null)
    navigate(`/compare/${idA}/${idB}`)
  }

  function goToVeto(event: React.FormEvent) {
    event.preventDefault()
    const id = sessionId.trim()
    if (!id) {
      setError('Veto: session id is required')
      return
    }
    setError(null)
    const query = team.trim() ? `?team=${encodeURIComponent(team.trim())}` : ''
    navigate(`/veto/${encodeURIComponent(id)}${query}`)
  }

  return (
    <nav className="border-b border-slate-200 bg-white px-4 py-2 text-sm">
      <div className="flex flex-wrap items-center gap-4">
        <span className="font-semibold text-slate-900">CS Analyser</span>

        <Link to="/" className="text-slate-600 hover:text-slate-900">
          Home
        </Link>
        <Link to="/admin" className="text-slate-600 hover:text-slate-900">
          Admin
        </Link>
        <Link to="/demos" className="text-slate-600 hover:text-slate-900">
          Demos
        </Link>
        {loggedIn ? (
          <button type="button" onClick={handleLogout} className="text-slate-600 hover:text-slate-900">
            Logout
          </button>
        ) : (
          <Link to="/login" className="text-slate-600 hover:text-slate-900">
            Login
          </Link>
        )}

        <form onSubmit={goToPlayer} className="flex items-center gap-1">
          <input
            value={playerId}
            onChange={(event) => setPlayerId(event.target.value)}
            placeholder="SteamID64 (numbers only)"
            className="w-44 rounded border border-slate-300 px-2 py-1 text-xs"
          />
          <button type="submit" className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700">
            Profile
          </button>
        </form>

        <form onSubmit={goToCompare} className="flex items-center gap-1">
          <input
            value={compareA}
            onChange={(event) => setCompareA(event.target.value)}
            placeholder="SteamID64 a"
            className="w-28 rounded border border-slate-300 px-2 py-1 text-xs"
          />
          <input
            value={compareB}
            onChange={(event) => setCompareB(event.target.value)}
            placeholder="SteamID64 b"
            className="w-28 rounded border border-slate-300 px-2 py-1 text-xs"
          />
          <button type="submit" className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700">
            Compare
          </button>
        </form>

        <form onSubmit={goToVeto} className="flex items-center gap-1">
          <input
            value={sessionId}
            onChange={(event) => setSessionId(event.target.value)}
            placeholder="veto session id"
            className="w-32 rounded border border-slate-300 px-2 py-1 text-xs"
          />
          <input
            value={team}
            onChange={(event) => setTeam(event.target.value)}
            placeholder="team (optional)"
            className="w-24 rounded border border-slate-300 px-2 py-1 text-xs"
          />
          <button type="submit" className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700">
            Veto
          </button>
        </form>
      </div>

      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
    </nav>
  )
}
