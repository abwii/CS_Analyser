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
