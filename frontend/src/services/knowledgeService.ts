import { apiGet, apiPut, apiDelete } from '@/utils/api'

export interface KnowledgePoint {
  id: number
  note_id: number
  content: string
  category?: string
  created_time: string
}

export interface UpdateKnowledgePointRequest {
  content?: string
  category?: string
}

export const getKnowledgePoints = () => {
  return apiGet<KnowledgePoint[]>('/knowledge-points')
}

export const getKnowledgePoint = (id: number) => {
  return apiGet<KnowledgePoint>(`/knowledge-points/${id}`)
}

export const updateKnowledgePoint = (id: number, data: UpdateKnowledgePointRequest) => {
  return apiPut<KnowledgePoint>(`/knowledge-points/${id}`, data)
}

export const deleteKnowledgePoint = (id: number) => {
  return apiDelete<void>(`/knowledge-points/${id}`)
}
