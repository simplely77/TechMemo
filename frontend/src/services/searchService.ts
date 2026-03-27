import { apiPost } from '@/utils/api'

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
  return apiPost<SemanticSearchResponse>('/search/semantic', data)
}