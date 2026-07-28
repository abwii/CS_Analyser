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
  z?: number
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

export interface RadarPlayerData {
  playerId: string
  playerName: string
  metrics: PlayerMetrics
}

export interface PlayerComparison {
  playerA: RadarPlayerData
  playerB: RadarPlayerData
}

export interface RadarAxisConfig {
  key: keyof PlayerMetrics
  label: string
  min: number
  max: number
}

export interface RadarPoint {
  x: number
  y: number
}

export interface LoginCredentials {
  email: string
  password?: string
}

export interface LoginResponse {
  token: string
  expiresAt: string
}

export interface FaceitMapStats {
  map: string
  matchesPlayed: number
  wins: number
  winRate: number
  avgKills: number
  avgDeaths: number
}

export interface FaceitStats {
  available: boolean
  faceitNickname?: string
  byMap: FaceitMapStats[]
}

export interface Demo {
  id: string
  mapName: string
  winner: string
  teamTScore: number
  teamCtScore: number
  playerCount: number
  createdAt: string
}

export interface UploadResult {
  bucket: string
  key: string
}

