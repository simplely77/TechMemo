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
  id: string
  title: string
  children?: MindMapNode[]
}

export interface GetMindMapResp {
  note_id: number
  root: MindMapNode
}

export interface GetGlobalMindMapResp {
  nodes: MindMapNode[]
  edges: Array<{ source: string; target: string }>
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
