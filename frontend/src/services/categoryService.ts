import request from '../utils/request'

export interface Category {
  id: number
  name: string
  description?: string
  created_time: string
}

export interface CreateCategoryRequest {
  name: string
  description?: string
}

export interface UpdateCategoryRequest {
  name?: string
  description?: string
}

export const getCategories = () => {
  return request.get<{ code: number; message: string; data: Category[] }>('/categories')
}

export const createCategory = (data: CreateCategoryRequest) => {
  return request.post<{ code: number; message: string; data: Category }>('/categories', data)
}

export const updateCategory = (id: number, data: UpdateCategoryRequest) => {
  return request.put<{ code: number; message: string; data: Category }>(`/categories/${id}`, data)
}

export const deleteCategory = (id: number) => {
  return request.delete<{ code: number; message: string }>(`/categories/${id}`)
}
