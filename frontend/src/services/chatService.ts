import { apiGet, apiPost } from '@/utils/api'

// ============ 类型定义 ============

export interface CreateSessionRequest {
  title?: string
}

export interface CreateSessionResp {
  id: number
  title: string
  created_at: string
  updated_at: string
}

export interface ChatSessionListResp {
  sessions: CreateSessionResp[]
  total: number
  page: number
  page_size: number
}

export interface SendMessageReq {
  content: string
}

export interface ChatMessageResp {
  id: number
  session_id: number
  role: string
  content: string
  created_at: string
}

export interface ChatMessageListResp {
  messages: ChatMessageResp[]
  total: number
  page: number
  page_size: number
}

// ============ API 调用 ============

/**
 * 创建聊天会话
 */
export const createSession = (data?: CreateSessionRequest) => {
  return apiPost<CreateSessionResp>('/chat/sessions', data || {})
}

/**
 * 获取会话列表
 */
export const getSessions = (page: number = 1, pageSize: number = 10) => {
  return apiGet<ChatSessionListResp>('/chat/sessions', {
    params: { page, page_size: pageSize }
  })
}

/**
 * 删除会话
 */
export const deleteSession = (id: number) => {
  return apiPost<void>(`/chat/sessions/${id}`, {}, { method: 'DELETE' })
}

/**
 * 发送消息
 */
export const sendMessage = (sessionId: number, data: SendMessageReq) => {
  return apiPost<ChatMessageResp>(`/chat/sessions/${sessionId}/messages`, data)
}

/**
 * 发送消息（流式返回）
 */
export const sendMessageStream = (sessionId: number, data: SendMessageReq) => {
  return fetch(`/api/v1/chat/sessions/${sessionId}/stream`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('access_token')}`
    },
    body: JSON.stringify(data)
  })
}

/**
 * 获取会话消息历史
 */
export const getMessages = (sessionId: number, page: number = 1, pageSize: number = 20) => {
  return apiGet<ChatMessageListResp>(`/chat/sessions/${sessionId}/messages`, {
    params: { page, page_size: pageSize }
  })
}
