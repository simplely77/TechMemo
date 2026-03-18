import request from '../utils/request'

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
  return request.get<{ code: number; message: string; data: Tag[] }>('/tags')
}

export const createTag = (data: CreateTagRequest) => {
  return request.post<{ code: number; message: string; data: Tag }>('/tags', data)
}

export const updateTag = (id: number, data: UpdateTagRequest) => {
  return request.put<{ code: number; message: string; data: Tag }>(`/tags/${id}`, data)
}

export const deleteTag = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/tags/${id}`)
}
