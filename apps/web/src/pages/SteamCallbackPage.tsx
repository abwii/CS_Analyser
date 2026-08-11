import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { getSteamId, setSteamToken } from '../lib/auth'

export function SteamCallbackPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const token = searchParams.get('token')
    if (!token) {
      setError('Steam login failed: no token received')
      return
    }

    setSteamToken(token)
    const steamId = getSteamId()
    if (!steamId) {
      setError('Steam login failed: could not read steamId from token')
      return
    }

    navigate(`/players/${steamId}`, { replace: true })
  }, [searchParams, navigate])

  if (error) {
    return <div className="p-8 text-sm text-red-600">{error}</div>
  }

  return <div className="p-8 text-sm text-slate-400">Signing in with Steam</div>
}
