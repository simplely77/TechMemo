import { apiGet, apiPost } from '@/utils/api'

export interface AIProcessResp {
  task_id: string
  status: string
}

export interface GetNoteAIStatusResp {
  note_id: number
  status: string
  created_time: string
}

export interface GetTaskStatusResp {
  task_id: string
  status: string
  result?: any
  error?: string
}

export interface MindMapNode {
  id: number
  name: string
  description: string
  importance_score: number
  children: MindMapNode[]
}

export interface GetMindMapResp {
  note_id: number
  nodes: MindMapNode[]
}

export interface GlobalMindMapNode {
  id: number
  note_id: number
  name: string
  description: string
  importance_score: number
}

export interface GlobalMindMapEdge {
  from: number
  to: number
  label: string
}

export interface GetGlobalMindMapResp {
  nodes: GlobalMindMapNode[]
  edges: GlobalMindMapEdge[]
}

export const processNoteAI = (noteId: number) => {
  return apiPost<AIProcessResp>(`/ai/note/${noteId}`, {})
}

export const getNoteAIStatus = (noteId: number) => {
  return apiGet<GetNoteAIStatusResp>(`/ai/note/${noteId}/status`)
}

export const generateGlobalMindMap = () => {
  return apiPost<AIProcessResp>('/ai/mindmap/global', {})
}

export const getTaskStatus = (taskId: string) => {
  return apiGet<GetTaskStatusResp>(`/ai/task/${taskId}/status`)
}

export const getMindMap = (noteId: number) => {
  return apiGet<GetMindMapResp>('/mindmap', { params: { note_id: noteId } })
}

export const getGlobalMindMap = () => {
  return apiGet<GetGlobalMindMapResp>('/mindmap/global')
}
