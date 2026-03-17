import request from '../utils/request'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user_id: number
  username: string
}

export interface RegisterResponse {
  user_id: number
  username: string
  created_time: string
}

export interface ProfileResponse {
  user_id: number
  username: string
  created_time: string
}

export const login = (data: LoginRequest) => {
  return request.post<{ code: number; message: string; data: LoginResponse }>('/auth/login', data)
}

export const register = (data: RegisterRequest) => {
  return request.post<{ code: number; message: string; data: RegisterResponse }>('/auth/register', data)
}

export const getProfile = () => {
  return request.get<{ code: number; message: string; data: ProfileResponse }>('/user/profile')
}