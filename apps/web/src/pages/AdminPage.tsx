import { withAuth } from '../hoc/withAuth'
import { UsersTable } from '../components/UsersTable'
import { QueueSizeIndicator } from '../components/QueueSizeIndicator'
import { MapsCrud } from '../components/MapsCrud'

function AdminPage() {
  return (
    <div className="min-h-screen bg-slate-50 p-8">
      <div className="mx-auto max-w-5xl space-y-8">
        <header className="flex items-center justify-between">
          <h1 className="text-xl font-semibold text-slate-900">Admin console</h1>
          <QueueSizeIndicator />
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
