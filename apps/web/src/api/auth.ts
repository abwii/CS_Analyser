import { API_BASE_URL, apiSend } from './client'
import type { LoginCredentials, LoginResponse } from '../types'

export function login(credentials: LoginCredentials): Promise<LoginResponse> {
  return apiSend<LoginResponse>('/auth/login', 'POST', credentials)
}

export function steamLoginUrl(): string {
  return `${API_BASE_URL}/auth/steam/login`
}
