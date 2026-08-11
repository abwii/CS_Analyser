export const TOKEN_STORAGE_KEY = 'cs_analyser_token'
export const STEAM_TOKEN_STORAGE_KEY = 'cs_analyser_steam_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY)
}

export function getSteamToken(): string | null {
  return localStorage.getItem(STEAM_TOKEN_STORAGE_KEY)
}

export function setSteamToken(token: string): void {
  localStorage.setItem(STEAM_TOKEN_STORAGE_KEY, token)
}

export function clearSteamToken(): void {
  localStorage.removeItem(STEAM_TOKEN_STORAGE_KEY)
}

export function getSteamId(): string | null {
  const token = getSteamToken()
  if (!isTokenValid(token)) {
    return null
  }
  const payload = decodeJwtPayload(token as string)
  const steamId = payload?.steamId
  return typeof steamId === 'string' ? steamId : null
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
