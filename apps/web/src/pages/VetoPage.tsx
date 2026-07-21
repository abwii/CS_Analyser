import { useParams, useSearchParams } from 'react-router-dom'
import { VetoBoard } from '../components/VetoBoard'
import { buildWebSocketUrl } from '../lib/ws'

export function VetoPage() {
  const { sessionId } = useParams<{ sessionId: string }>()
  const [searchParams] = useSearchParams()
  const localTeam = searchParams.get('team')

  if (!sessionId) {
    return <div className="p-8 text-sm text-red-600">Missing session id</div>
  }

  const wsUrl = buildWebSocketUrl(`/veto/ws?sessionId=${sessionId}`)

  return <VetoBoard wsUrl={wsUrl} localTeam={localTeam} />
}
