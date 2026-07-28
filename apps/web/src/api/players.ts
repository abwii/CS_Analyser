import { apiGet } from './client'
import type { FaceitStats, PlayerProfile } from '../types'

export function fetchPlayerProfile(playerId: string): Promise<PlayerProfile> {
  return apiGet<PlayerProfile>(`/players/${playerId}/profile`)
}

export function fetchFaceitStats(playerId: string): Promise<FaceitStats> {
  return apiGet<FaceitStats>(`/players/${playerId}/faceit-stats`)
}
