import { useCallback, useEffect, useRef, useState } from 'react'
import { fetchDemos, uploadDemo } from '../api/demos'
import { withAuth } from '../hoc/withAuth'
import type { Demo } from '../types'

const pollIntervalMs = 3000
const pollAttempts = 10

interface DemoGroup {
  mapName: string
  demos: Demo[]
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

function groupByMap(demos: Demo[]): DemoGroup[] {
  const groups = new Map<string, Demo[]>()
  for (const demo of demos) {
    const key = demo.mapName || 'Unknown map'
    const bucket = groups.get(key)
    if (bucket) {
      bucket.push(demo)
    } else {
      groups.set(key, [demo])
    }
  }

  const result: DemoGroup[] = Array.from(groups.entries()).map(([mapName, mapDemos]) => ({
    mapName,
    demos: [...mapDemos].sort((a, b) => b.createdAt.localeCompare(a.createdAt)),
  }))

  result.sort((a, b) => b.demos[0].createdAt.localeCompare(a.demos[0].createdAt))
  return result
}

function DemosPage() {
  const [demos, setDemos] = useState<Demo[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [dragActive, setDragActive] = useState(false)
  const [uploadStatus, setUploadStatus] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const pollTimeoutRef = useRef<number | null>(null)

  const loadDemos = useCallback(() => {
    return fetchDemos()
      .then((data) => {
        setDemos(data)
        setError(null)
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : 'failed to load demos')
      })
  }, [])

  useEffect(() => {
    loadDemos().finally(() => setLoading(false))
    return () => {
      if (pollTimeoutRef.current !== null) {
        window.clearTimeout(pollTimeoutRef.current)
      }
    }
  }, [loadDemos])

  function pollForProcessing(remaining: number) {
    if (remaining <= 0) {
      return
    }
    pollTimeoutRef.current = window.setTimeout(() => {
      void loadDemos().then(() => pollForProcessing(remaining - 1))
    }, pollIntervalMs)
  }

  async function handleFiles(files: FileList | null) {
    if (!files || files.length === 0) {
      return
    }
    const file = files[0]
    setUploadStatus(`Uploading ${file.name}`)
    try {
      await uploadDemo(file)
      setUploadStatus(`${file.name} uploaded, waiting for processing`)
      pollForProcessing(pollAttempts)
    } catch (err) {
      setUploadStatus(err instanceof Error ? err.message : 'upload failed')
    }
  }

  function handleDrop(event: React.DragEvent<HTMLDivElement>) {
    event.preventDefault()
    setDragActive(false)
    void handleFiles(event.dataTransfer.files)
  }

  function handleDragOver(event: React.DragEvent<HTMLDivElement>) {
    event.preventDefault()
    setDragActive(true)
  }

  function handleDragLeave() {
    setDragActive(false)
  }

  const groups = groupByMap(demos)

  return (
    <div className="min-h-screen bg-slate-50 p-8">
      <div className="mx-auto max-w-4xl space-y-6">
        <header>
          <h1 className="text-xl font-semibold text-slate-900">Demos</h1>
          <p className="text-sm text-slate-500">Upload demo files and browse processed matches</p>
        </header>

        <div
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onClick={() => fileInputRef.current?.click()}
          className={`flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed px-4 py-10 text-center transition-colors ${
            dragActive ? 'border-slate-500 bg-slate-100' : 'border-slate-300 bg-white'
          }`}
        >
          <p className="text-sm font-medium text-slate-700">Drag and drop a demo file here</p>
          <p className="mt-1 text-xs text-slate-400">or click to browse (.dem)</p>
          <input
            ref={fileInputRef}
            type="file"
            accept=".dem"
            className="hidden"
            onChange={(event) => void handleFiles(event.target.files)}
          />
        </div>

        {uploadStatus && <p className="text-sm text-slate-600">{uploadStatus}</p>}

        <div className="rounded-lg border border-slate-200 bg-white">
          <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Processed demos</h2>
            <button
              type="button"
              onClick={() => void loadDemos()}
              className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700"
            >
              Refresh
            </button>
          </div>

          {error && <p className="px-4 py-3 text-sm text-red-600">{error}</p>}

          {!error && loading && <p className="px-4 py-3 text-sm text-slate-400">Loading demos</p>}

          {!error && !loading && demos.length === 0 && (
            <p className="px-4 py-3 text-sm text-slate-400">No processed demos yet</p>
          )}

          {!error &&
            groups.map((group) => (
              <div key={group.mapName} className="border-b border-slate-100 last:border-0">
                <div className="bg-slate-50 px-4 py-2 text-sm font-medium text-slate-700">
                  {group.mapName} ({group.demos.length})
                </div>
                <table className="w-full text-left text-sm">
                  <thead className="border-b border-slate-100 text-slate-500">
                    <tr>
                      <th className="px-4 py-2 font-medium">Date</th>
                      <th className="px-4 py-2 font-medium">Score</th>
                      <th className="px-4 py-2 font-medium">Winner</th>
                      <th className="px-4 py-2 font-medium">Players</th>
                    </tr>
                  </thead>
                  <tbody>
                    {group.demos.map((demo) => (
                      <tr key={demo.id} className="border-b border-slate-100 last:border-0">
                        <td className="px-4 py-2 text-slate-600">{formatDate(demo.createdAt)}</td>
                        <td className="px-4 py-2 text-slate-600">
                          {demo.teamTScore} - {demo.teamCtScore}
                        </td>
                        <td className="px-4 py-2 text-slate-600">{demo.winner || 'Unknown'}</td>
                        <td className="px-4 py-2 text-slate-600">{demo.playerCount}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ))}
        </div>
      </div>
    </div>
  )
}

const ProtectedDemosPage = withAuth(DemosPage)

export default ProtectedDemosPage
