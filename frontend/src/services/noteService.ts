import request from '../utils/request'

export interface Note {
  id: number
  title: string
  content: string
  category_id?: number
  status: string
  created_time: string
  updated_time: string
}

export interface CreateNoteRequest {
  title: string
  content: string
  category_id?: number
}

export interface UpdateNoteRequest {
  title?: string
  content?: string
  category_id?: number
}

export interface NoteVersion {
  version_id: number
  content: string
  created_time: string
}

export const getNotes = () => {
  return request.get<{ code: number; message: string; data: Note[] }>('/notes')
}

export const createNote = (data: CreateNoteRequest) => {
  return request.post<{ code: number; message: string; data: Note }>('/notes', data)
}

export const getNote = (id: number) => {
  return request.get<{ code: number; message: string; data: Note }>(`/notes/${id}`)
}

export const updateNote = (id: number, data: UpdateNoteRequest) => {
  return request.put<{ code: number; message: string; data: Note }>(`/notes/${id}`, data)
}

export const updateNoteTags = (id: number, tagIds: number[]) => {
  return request.put<{ code: number; message: string; data: Note }>(`/notes/${id}/tags`, { tag_ids: tagIds })
}

export const deleteNote = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/notes/${id}`)
}

export const getNoteVersions = (id: number) => {
  return request.get<{ code: number; message: string; data: NoteVersion[] }>(`/notes/${id}/versions`)
}

export const restoreNote = (id: number, versionId: number) => {
  return request.post<{ code: number; message: string; data: Note }>(`/notes/${id}/versions/${versionId}/restore`, {})
}
