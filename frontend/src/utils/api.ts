import request from './request'
import type { ApiResponse } from '@/types/response'
import { ApiError } from '@/types/response'
import { toast } from 'sonner'

/**
 * 处理 API 响应
 */
const handleResponse = async <T>(promise: Promise<any>): Promise<T> => {
  try {
    const response = await promise
    const { code, message, data } = response.data as ApiResponse<T>

    if (code !== 20000) {
      const errorMsg = message || '请求失败'
      toast.error(errorMsg)
      throw new ApiError(code, errorMsg)
    }

    return data
  } catch (error: any) {
    if (error instanceof ApiError) {
      throw error
    }

    if (error.response?.data) {
      const { code, message } = error.response.data
      const errorMsg = message || '请求失败'
      toast.error(errorMsg)
      throw new ApiError(code, errorMsg, error)
    }

    const errorMsg = error.message || '请求失败'
    toast.error(errorMsg)
    throw new ApiError(-1, errorMsg, error)
  }
}

/**
 * GET 请求
 */
export const apiGet = <T>(url: string, config?: any): Promise<T> => {
  return handleResponse<T>(request.get(url, config))
}

/**
 * POST 请求
 */
export const apiPost = <T>(url: string, data?: any, config?: any): Promise<T> => {
  return handleResponse<T>(request.post(url, data, config))
}

/**
 * PUT 请求
 */
export const apiPut = <T>(url: string, data?: any, config?: any): Promise<T> => {
  return handleResponse<T>(request.put(url, data, config))
}

/**
 * DELETE 请求
 */
export const apiDelete = <T>(url: string, config?: any): Promise<T> => {
  return handleResponse<T>(request.delete(url, config))
}
