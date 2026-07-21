import { apiDelete, apiGet, apiSend } from './client'
import type { CompetitiveMap } from '../types'

export interface MapInput {
  name: string
  active: boolean
}

export function fetchMaps(): Promise<CompetitiveMap[]> {
  return apiGet<CompetitiveMap[]>('/admin/maps')
}

export function createMap(input: MapInput): Promise<CompetitiveMap> {
  return apiSend<CompetitiveMap>('/admin/maps', 'POST', input)
}

export function updateMap(id: string, input: MapInput): Promise<CompetitiveMap> {
  return apiSend<CompetitiveMap>(`/admin/maps/${id}`, 'PUT', input)
}

export function deleteMap(id: string): Promise<void> {
  return apiDelete(`/admin/maps/${id}`)
}
