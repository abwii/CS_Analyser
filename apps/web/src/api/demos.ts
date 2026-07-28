import { apiGet, apiUpload } from './client'
import type { Demo, UploadResult } from '../types'

export function fetchDemos(): Promise<Demo[]> {
  return apiGet<Demo[]>('/demos')
}

export function uploadDemo(file: File): Promise<UploadResult> {
  return apiUpload<UploadResult>('/upload', file)
}
