import { apiSend } from './client'
import type { LoginCredentials, LoginResponse } from '../types'

export function login(credentials: LoginCredentials): Promise<LoginResponse> {
  return apiSend<LoginResponse>('/auth/login', 'POST', credentials)
}
