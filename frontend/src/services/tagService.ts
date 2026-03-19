import { apiGet, apiPost, apiPut, apiDelete } from '@/utils/api'

export interface Tag {
  id: number
  name: string
  description?: string
  created_time: string
}

export interface CreateTagRequest {
  name: string
  description?: string
}

export interface UpdateTagRequest {
  name?: string
  description?: string
}

export const getTags = () => {
  return apiGet<Tag[]>('/tags')
}

export const createTag = (data: CreateTagRequest) => {
  return apiPost<Tag>('/tags', data)
}

export const updateTag = (id: number, data: UpdateTagRequest) => {
  return apiPut<Tag>(`/tags/${id}`, data)
}

export const deleteTag = (id: number) => {
  return apiDelete<void>(`/tags/${id}`)
}
