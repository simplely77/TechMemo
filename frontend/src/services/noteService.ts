import { apiGet, apiPost, apiPut, apiDelete } from '@/utils/api'

// ============ 类型定义 ============

export interface NoteTag {
  id: number
  name: string
}

export interface NoteCategory {
  id: number
  name: string
}

export interface Note {
  id: number
  title: string
  content_md: string
  note_type?: string
  category: NoteCategory
  tags: NoteTag[]
  status: string
  created_at: string
  updated_at: string
}

export interface CreateNoteRequest {
  title: string
  content_md: string
  category_id: number
  tag_ids?: number[]
}

export interface UpdateNoteRequest {
  title?: string
  content_md?: string
  category_id?: number
}

export interface NoteVersion {
  id: number
  note_id: number
  content_md: string
  created_at: string
}

export interface GetNoteVersionsResponse {
  versions: NoteVersion[]
}

export interface GetNotesQuery {
  category_id?: number
  tag_ids?: number[]
  keyword?: string
  note_type?: string
  page?: number
  page_size?: number
  sort?: string
}

export interface GetNotesResponse {
  notes?: Note[]
  total: number
  page: number
  page_size: number
}

// ============ API 调用 ============

/**
 * 获取笔记列表
 */
export const getNotes = (query?: GetNotesQuery) => {
  return apiGet<GetNotesResponse>('/notes', { params: query })
}

/**
 * 创建笔记
 */
export const createNote = (data: CreateNoteRequest) => {
  return apiPost<Note>('/notes', data)
}

/**
 * 获取笔记详情
 */
export const getNote = (id: number) => {
  return apiGet<Note>(`/notes/${id}`)
}

/**
 * 更新笔记
 */
export const updateNote = (id: number, data: UpdateNoteRequest) => {
  return apiPut<Note>(`/notes/${id}`, data)
}

/**
 * 更新笔记标签
 */
export const updateNoteTags = (id: number, tagIds: number[]) => {
  return apiPut<Note>(`/notes/${id}/tags`, { tag_ids: tagIds })
}

/**
 * 删除笔记
 */
export const deleteNote = (id: number) => {
  return apiDelete<void>(`/notes/${id}`)
}

/**
 * 获取笔记版本历史
 */
export const getNoteVersions = (id: number, sort: string = 'created_at_desc') => {
  return apiGet<GetNoteVersionsResponse>(`/notes/${id}/versions`, { params: { sort } })
}

/**
 * 恢复笔记到指定版本
 */
export const restoreNote = (id: number, versionId: number) => {
  return apiPost<Note>(`/notes/${id}/versions/${versionId}/restore`, {})
}
