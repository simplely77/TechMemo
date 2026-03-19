import { apiGet, apiPost, apiPut, apiDelete } from '@/utils/api'

export interface Category {
  id: number
  name: string
  created_time: string
}

export interface CreateCategoryRequest {
  name: string
}

export interface UpdateCategoryRequest {
  name: string
}

export interface GetCategoriesResp {
  categories: Category[]
}

export const getCategories = () => {
  return apiGet<GetCategoriesResp>('/categories')
}

export const createCategory = (data: CreateCategoryRequest) => {
  return apiPost<Category>('/categories', data)
}

export const updateCategory = (id: number, data: UpdateCategoryRequest) => {
  return apiPut<Category>(`/categories/${id}`, data)
}

export const deleteCategory = (id: number) => {
  return apiDelete<void>(`/categories/${id}`)
}
