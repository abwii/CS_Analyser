import { useEffect, useState } from 'react'
import { createMap, deleteMap, fetchMaps, updateMap } from '../api/maps'
import type { CompetitiveMap } from '../types'

interface EditState {
  id: string
  name: string
  active: boolean
}

export function MapsCrud() {
  const [maps, setMaps] = useState<CompetitiveMap[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [newName, setNewName] = useState('')
  const [newActive, setNewActive] = useState(true)

  const [editing, setEditing] = useState<EditState | null>(null)

  function loadMaps() {
    setLoading(true)
    setError(null)
    fetchMaps()
      .then(setMaps)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'failed to load maps'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    loadMaps()
  }, [])

  function handleCreate(event: React.FormEvent) {
    event.preventDefault()
    if (!newName.trim()) {
      return
    }
    createMap({ name: newName.trim(), active: newActive })
      .then(() => {
        setNewName('')
        setNewActive(true)
        loadMaps()
      })
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'failed to create map'))
  }

  function startEdit(map: CompetitiveMap) {
    setEditing({ id: map.id, name: map.name, active: map.active })
  }

  function handleSaveEdit() {
    if (!editing) {
      return
    }
    updateMap(editing.id, { name: editing.name.trim(), active: editing.active })
      .then(() => {
        setEditing(null)
        loadMaps()
      })
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'failed to update map'))
  }

  function handleDelete(id: string) {
    if (!window.confirm('Delete this map')) {
      return
    }
    deleteMap(id)
      .then(loadMaps)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'failed to delete map'))
  }

  return (
    <div className="rounded-lg border border-slate-200 bg-white">
      <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-3 border-b border-slate-200 p-4">
        <div className="flex flex-col">
          <label htmlFor="new-map-name" className="text-xs font-medium text-slate-500">
            Map name
          </label>
          <input
            id="new-map-name"
            type="text"
            value={newName}
            onChange={(event) => setNewName(event.target.value)}
            className="mt-1 rounded border border-slate-300 px-3 py-1.5 text-sm"
            placeholder="de_mirage"
          />
        </div>
        <label className="flex items-center gap-2 pb-1.5 text-sm text-slate-600">
          <input
            type="checkbox"
            checked={newActive}
            onChange={(event) => setNewActive(event.target.checked)}
            className="h-4 w-4 rounded border-slate-300"
          />
          Active
        </label>
        <button
          type="submit"
          className="rounded bg-slate-900 px-4 py-1.5 text-sm font-medium text-white disabled:opacity-40"
          disabled={!newName.trim()}
        >
          Add map
        </button>
      </form>

      {loading && <p className="px-4 py-3 text-sm text-slate-400">Loading</p>}
      {error && <p className="px-4 py-3 text-sm text-red-600">{error}</p>}

      <ul>
        {maps.map((map) => (
          <li key={map.id} className="flex items-center justify-between gap-3 border-b border-slate-100 px-4 py-3 last:border-0">
            {editing?.id === map.id ? (
              <>
                <div className="flex flex-1 items-center gap-3">
                  <input
                    type="text"
                    value={editing.name}
                    onChange={(event) => setEditing({ ...editing, name: event.target.value })}
                    className="rounded border border-slate-300 px-3 py-1.5 text-sm"
                  />
                  <label className="flex items-center gap-2 text-sm text-slate-600">
                    <input
                      type="checkbox"
                      checked={editing.active}
                      onChange={(event) => setEditing({ ...editing, active: event.target.checked })}
                      className="h-4 w-4 rounded border-slate-300"
                    />
                    Active
                  </label>
                </div>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={handleSaveEdit}
                    className="rounded bg-slate-900 px-3 py-1 text-sm font-medium text-white"
                  >
                    Save
                  </button>
                  <button
                    type="button"
                    onClick={() => setEditing(null)}
                    className="rounded border border-slate-300 px-3 py-1 text-sm text-slate-700"
                  >
                    Cancel
                  </button>
                </div>
              </>
            ) : (
              <>
                <div className="flex items-center gap-3">
                  <span className={`h-2.5 w-2.5 rounded-full ${map.active ? 'bg-emerald-500' : 'bg-slate-300'}`} />
                  <span className="text-sm text-slate-800">{map.name}</span>
                </div>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => startEdit(map)}
                    className="rounded border border-slate-300 px-3 py-1 text-sm text-slate-700"
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(map.id)}
                    className="rounded border border-red-200 px-3 py-1 text-sm text-red-600"
                  >
                    Delete
                  </button>
                </div>
              </>
            )}
          </li>
        ))}
        {!loading && maps.length === 0 && <li className="px-4 py-6 text-center text-sm text-slate-400">No maps configured</li>}
      </ul>
    </div>
  )
}
