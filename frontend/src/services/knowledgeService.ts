import request from '../utils/request'

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
  return request.get<{ code: number; message: string; data: KnowledgePoint[] }>('/knowledge-points')
}

export const getKnowledgePoint = (id: number) => {
  return request.get<{ code: number; message: string; data: KnowledgePoint }>(`/knowledge-points/${id}`)
}

export const updateKnowledgePoint = (id: number, data: UpdateKnowledgePointRequest) => {
  return request.put<{ code: number; message: string; data: KnowledgePoint }>(`/knowledge-points/${id}`, data)
}

export const deleteKnowledgePoint = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/knowledge-points/${id}`)
}
