import { apiGet } from './client'
import type { FaceitRecentMatches, FaceitStats, PlayerProfile, SteamSummary } from '../types'

export function fetchPlayerProfile(playerId: string): Promise<PlayerProfile> {
  return apiGet<PlayerProfile>(`/players/${playerId}/profile`)
}

export function fetchFaceitStats(playerId: string): Promise<FaceitStats> {
  return apiGet<FaceitStats>(`/players/${playerId}/faceit-stats`)
}

export function fetchFaceitRecentMatches(playerId: string): Promise<FaceitRecentMatches> {
  return apiGet<FaceitRecentMatches>(`/players/${playerId}/faceit-matches`)
}

export function fetchSteamSummary(playerId: string): Promise<SteamSummary> {
  return apiGet<SteamSummary>(`/players/${playerId}/steam-summary`)
}
