import { apiGet, apiPut, apiDelete } from '@/utils/api'

export interface KnowledgePoint {
  id: number
  name: string
  description: string
  source_note_id: number
  source_note_title?: string
  importance_score: number
  created_at: string
}

export interface GetKnowledgePointsResponse {
  knowledge_points: KnowledgePoint[]
  total: number
  page: number
  page_size: number
}

export interface UpdateKnowledgePointRequest {
  name?: string
  description?: string
  importance_score?: number
}

export const getKnowledgePoints = (params?: {
  source_note_id?: number
  keyword?: string
  min_importance?: number
  page?: number
  page_size?: number
}) => {
  return apiGet<GetKnowledgePointsResponse>('/knowledge-points', { params })
}

export interface GetKnowledgePointResponse {
  id: number
  name: string
  description: string
  source_note_id: number
  source_note_title?: string
  importance_score: number
  created_at: string
  related_points?: Array<{
    id: number
    name: string
    relation_type: string
  }>
}

export const getKnowledgePoint = (id: number) => {
  return apiGet<GetKnowledgePointResponse>(`/knowledge-points/${id}`)
}

export const updateKnowledgePoint = (id: number, data: UpdateKnowledgePointRequest) => {
  return apiPut<KnowledgePoint>(`/knowledge-points/${id}`, data)
}

export const deleteKnowledgePoint = (id: number) => {
  return apiDelete<void>(`/knowledge-points/${id}`)
}
