export interface User {
  id: string
  email: string
  role: string
  createdAt: string
}

export interface PaginatedUsers {
  users: User[]
  total: number
  page: number
  pageSize: number
}

export interface CompetitiveMap {
  id: string
  name: string
  active: boolean
}

export interface QueueSize {
  size: number
}

export type VetoFormat = 'BO1' | 'BO3'

export type VetoActionType = 'ban' | 'pick'

export interface VetoEntry {
  team: string
  map: string
  action: VetoActionType
}

export interface VetoState {
  sessionId: string
  format: VetoFormat
  teamA: string
  teamB: string
  availableMaps: string[]
  currentTurn: string
  nextAction?: VetoActionType
  history: VetoEntry[]
  deciderMap?: string
  completed: boolean
}

export interface Point2D {
  x: number
  y: number
}

export interface WorldBounds {
  minX: number
  maxX: number
  minY: number
  maxY: number
}

export interface CanvasSize {
  width: number
  height: number
}

export interface PlayerMetrics {
  kpr: number
  dpr: number
  kast: number
  impact: number
  adr: number
}

export interface PlayerMap {
  name: string
  imageUrl: string
  bounds: WorldBounds
}

export interface PlayerProfile {
  playerId: string
  playerName: string
  metrics: PlayerMetrics
  map: PlayerMap
  positions: Point2D[]
}

