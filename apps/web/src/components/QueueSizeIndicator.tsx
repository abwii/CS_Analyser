import { useEffect, useState } from 'react'
import { fetchQueueSize } from '../api/queue'

const REFRESH_INTERVAL_MS = 10000
const WARNING_THRESHOLD = 50
const CRITICAL_THRESHOLD = 200

function levelForSize(size: number): 'ok' | 'warning' | 'critical' {
  if (size >= CRITICAL_THRESHOLD) {
    return 'critical'
  }
  if (size >= WARNING_THRESHOLD) {
    return 'warning'
  }
  return 'ok'
}

const LEVEL_STYLES: Record<'ok' | 'warning' | 'critical', string> = {
  ok: 'bg-emerald-500',
  warning: 'bg-amber-500',
  critical: 'bg-red-500',
}

export function QueueSizeIndicator() {
  const [size, setSize] = useState<number | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    function load() {
      fetchQueueSize()
        .then((data) => {
          if (!cancelled) {
            setSize(data.size)
            setError(null)
          }
        })
        .catch((err: unknown) => {
          if (!cancelled) {
            setError(err instanceof Error ? err.message : 'failed to load queue size')
          }
        })
    }

    load()
    const interval = setInterval(load, REFRESH_INTERVAL_MS)

    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [])

  if (error) {
    return <span className="text-sm text-red-600">{error}</span>
  }

  if (size === null) {
    return <span className="text-sm text-slate-400">Loading queue size</span>
  }

  const level = levelForSize(size)

  return (
    <div className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-3 py-1.5">
      <span className={`h-2.5 w-2.5 rounded-full ${LEVEL_STYLES[level]}`} />
      <span className="text-sm text-slate-700">SQS queue</span>
      <span className="text-sm font-semibold text-slate-900">{size}</span>
    </div>
  )
}
