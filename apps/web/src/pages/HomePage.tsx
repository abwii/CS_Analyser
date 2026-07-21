import { Link } from 'react-router-dom'

export function HomePage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50">
      <div className="text-center">
        <h1 className="text-2xl font-semibold text-slate-900">CS Analyser</h1>
        <Link to="/admin" className="mt-4 inline-block text-sm text-slate-600 underline">
          Admin console
        </Link>
      </div>
    </div>
  )
}
