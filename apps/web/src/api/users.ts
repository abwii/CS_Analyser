import { apiGet } from './client'
import type { PaginatedUsers } from '../types'

export function fetchUsers(page: number, pageSize: number): Promise<PaginatedUsers> {
  return apiGet<PaginatedUsers>(`/admin/users?page=${page}&pageSize=${pageSize}`)
}
