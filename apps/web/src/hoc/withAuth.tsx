import type { ComponentType } from 'react'
import { Navigate } from 'react-router-dom'
import { getToken, isTokenValid } from '../lib/auth'

export function withAuth<P extends object>(Component: ComponentType<P>): ComponentType<P> {
  return function ProtectedComponent(props: P) {
    const token = getToken()
    if (!isTokenValid(token)) {
      return <Navigate to="/login" replace />
    }
    return <Component {...props} />
  }
}
