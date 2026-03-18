import request from '../utils/request'

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
  return request.post<{ code: number; message: string; data: AIProcessResp }>(`/ai/note/${noteId}`, {})
}

export const getNoteAIStatus = (noteId: number) => {
  return request.get<{ code: number; message: string; data: GetNoteAIStatusResp }>(`/ai/note/${noteId}/status`)
}

export const generateGlobalMindMap = () => {
  return request.post<{ code: number; message: string; data: AIProcessResp }>('/ai/mindmap/global', {})
}

export const getTaskStatus = (taskId: string) => {
  return request.get<{ code: number; message: string; data: GetTaskStatusResp }>(`/ai/task/${taskId}/status`)
}

export const getMindMap = (noteId: number) => {
  return request.get<{ code: number; message: string; data: GetMindMapResp }>('/mindmap', { params: { note_id: noteId } })
}

export const getGlobalMindMap = () => {
  return request.get<{ code: number; message: string; data: GetGlobalMindMapResp }>('/mindmap/global')
}
