import { getToken } from '../lib/auth'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

function authHeaders(): HeadersInit {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: authHeaders(),
  })
  if (!response.ok) {
    throw new Error(`request to ${path} failed with status ${response.status}`)
  }
  return response.json() as Promise<T>
}

export async function apiSend<T>(path: string, method: string, body: unknown): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers: {
      ...authHeaders(),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })
  if (!response.ok) {
    throw new Error(`request to ${path} failed with status ${response.status}`)
  }
  return response.json() as Promise<T>
}

export async function apiDelete(path: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: 'DELETE',
    headers: authHeaders(),
  })
  if (!response.ok) {
    throw new Error(`request to ${path} failed with status ${response.status}`)
  }
}
