export function buildWebSocketUrl(path: string): string {
  const httpBase = import.meta.env.VITE_API_BASE_URL || window.location.origin
  const wsBase = httpBase.replace(/^http/, 'ws')
  return `${wsBase}${path}`
}
