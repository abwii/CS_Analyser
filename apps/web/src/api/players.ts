import { apiGet } from './client'
import type { PlayerProfile } from '../types'

export function fetchPlayerProfile(playerId: string): Promise<PlayerProfile> {
  return apiGet<PlayerProfile>(`/players/${playerId}/profile`)
}
