import { apiGet } from './client'
import type { QueueSize } from '../types'

export function fetchQueueSize(): Promise<QueueSize> {
  return apiGet<QueueSize>('/admin/queue-size')
}
