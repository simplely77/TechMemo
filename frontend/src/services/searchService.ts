import request from '../utils/request'

export interface SearchResult {
  id: number
  type: string
  title: string
  content: string
  similarity: number
}

export interface SemanticSearchRequest {
  query: string
  search_type?: string
  top_k?: number
}

export interface SemanticSearchResponse {
  results: SearchResult[]
  total: number
}

export const semanticSearch = (data: SemanticSearchRequest) => {
  return request.post<{ code: number; message: string; data: SemanticSearchResponse }>('/search/semantic', data)
}

export const askQuestion = (question: string) => {
  return request.post<{ code: number; message: string; data: any }>('/qa/ask', { question })
}
