import { useNavigate } from 'react-router-dom'
import { clearToken } from '../lib/auth'
import { withAuth } from '../hoc/withAuth'
import { UsersTable } from '../components/UsersTable'
import { QueueSizeIndicator } from '../components/QueueSizeIndicator'
import { MapsCrud } from '../components/MapsCrud'

function AdminPage() {
  const navigate = useNavigate()

  const handleLogout = () => {
    clearToken()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-screen bg-slate-50 p-8">
      <div className="mx-auto max-w-5xl space-y-8">
        <header className="flex items-center justify-between">
          <h1 className="text-xl font-semibold text-slate-900">Admin console</h1>
          <div className="flex items-center gap-4">
            <QueueSizeIndicator />
            <button
              onClick={handleLogout}
              className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-500"
            >
              Logout
            </button>
          </div>
        </header>

        <section className="space-y-3">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Users</h2>
          <UsersTable />
        </section>

        <section className="space-y-3">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Competitive maps</h2>
          <MapsCrud />
        </section>
      </div>
    </div>
  )
}

const ProtectedAdminPage = withAuth(AdminPage)

export default ProtectedAdminPage
