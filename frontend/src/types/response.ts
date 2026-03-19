/**
 * 通用 API 响应类型
 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

/**
 * API 错误类
 */
export class ApiError extends Error {
  code: number
  message: string
  originalError?: any

  constructor(code: number, message: string, originalError?: any) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.message = message
    this.originalError = originalError
  }
}
