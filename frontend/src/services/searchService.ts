import { apiPost } from '@/utils/api'

export interface SearchResult {
  id: number
  type: string
  title: string
  content: string
  similarity: number
  created_at: string
  // Note 特有字段
  note_type?: string
  category?: {
    id: number
    name: string
  }
  // Knowledge 特有字段
  source_note_id?: number
  source_note_title?: string
  importance_score?: number
}

export interface SemanticSearchRequest {
  query: string
  search_type: string  // 必需：'note' 或 'knowledge'
  top_k: number        // 必需：1-20
}

export interface SemanticSearchResponse {
  results: SearchResult[]
  query: string
  total: number
}

export const semanticSearch = (data: SemanticSearchRequest) => {
  return apiPost<SemanticSearchResponse>('/search/semantic', data)
}