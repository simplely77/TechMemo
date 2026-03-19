import { apiGet, apiPost } from '@/utils/api'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
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

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface RefreshTokenResponse {
  access_token: string
  refresh_token: string
}

export const login = (data: LoginRequest) => {
  return apiPost<LoginResponse>('/auth/login', data)
}

export const register = (data: RegisterRequest) => {
  return apiPost<RegisterResponse>('/auth/register', data)
}

export const getProfile = () => {
  return apiGet<ProfileResponse>('/auth/profile')
}

export const refreshToken = (data: RefreshTokenRequest) => {
  return apiPost<RefreshTokenResponse>('/auth/refresh', data)
}