import { useEffect, useState } from 'react'
import { fetchUsers } from '../api/users'
import type { User } from '../types'

const PAGE_SIZE = 10

export function UsersTable() {
  const [users, setUsers] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    fetchUsers(page, PAGE_SIZE)
      .then((data) => {
        if (cancelled) {
          return
        }
        setUsers(data.users)
        setTotal(data.total)
      })
      .catch((err: unknown) => {
        if (cancelled) {
          return
        }
        setError(err instanceof Error ? err.message : 'failed to load users')
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [page])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className="rounded-lg border border-slate-200 bg-white">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-slate-200 text-slate-500">
            <tr>
              <th className="px-4 py-3 font-medium">Email</th>
              <th className="px-4 py-3 font-medium">Role</th>
              <th className="px-4 py-3 font-medium">Created</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <tr key={user.id} className="border-b border-slate-100 last:border-0">
                <td className="px-4 py-3 text-slate-800">{user.email}</td>
                <td className="px-4 py-3 text-slate-600">{user.role}</td>
                <td className="px-4 py-3 text-slate-600">{new Date(user.createdAt).toLocaleDateString()}</td>
              </tr>
            ))}
            {!loading && users.length === 0 && (
              <tr>
                <td className="px-4 py-6 text-center text-slate-400" colSpan={3}>
                  No users found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {loading && <p className="px-4 py-3 text-sm text-slate-400">Loading</p>}
      {error && <p className="px-4 py-3 text-sm text-red-600">{error}</p>}

      <div className="flex items-center justify-between border-t border-slate-200 px-4 py-3">
        <span className="text-sm text-slate-500">
          Page {page} of {totalPages}
        </span>
        <div className="flex gap-2">
          <button
            type="button"
            className="rounded border border-slate-300 px-3 py-1 text-sm text-slate-700 disabled:opacity-40"
            disabled={page <= 1}
            onClick={() => setPage((current) => current - 1)}
          >
            Previous
          </button>
          <button
            type="button"
            className="rounded border border-slate-300 px-3 py-1 text-sm text-slate-700 disabled:opacity-40"
            disabled={page >= totalPages}
            onClick={() => setPage((current) => current + 1)}
          >
            Next
          </button>
        </div>
      </div>
    </div>
  )
}
