export const TOKEN_STORAGE_KEY = 'cs_analyser_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY)
}

function decodeJwtPayload(token: string): Record<string, unknown> | null {
  const parts = token.split('.')
  if (parts.length !== 3) {
    return null
  }
  try {
    const normalized = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const decoded = atob(normalized)
    return JSON.parse(decoded) as Record<string, unknown>
  } catch {
    return null
  }
}

export function isTokenValid(token: string | null): boolean {
  if (!token) {
    return false
  }
  const payload = decodeJwtPayload(token)
  if (!payload) {
    return false
  }
  const expiresAt = payload.exp
  if (typeof expiresAt !== 'number') {
    return true
  }
  return Date.now() < expiresAt * 1000
}
