import { apiPost, apiGet } from '@/utils/api'

export interface SearchResult {
  id: number
  type: string
  title: string
  content: string
  similarity: number
  /** CrossEncoder 原始分，仅混合检索且 rerank 成功时存在 */
  rerank_score?: number
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
  top_k: number        // 必需：与后端一致，1–100
}

export interface SemanticSearchResponse {
  results: SearchResult[]
  query: string
  total: number
}
export interface SearchHistoryItem {
  id: number
  keyword: string
  search_type: string
  target_type: string
  last_searched_at: string
  created_at: string
}

export interface GetSearchHistory{
  page: number
  page_size: number
  search_type: string
  target_type: string
}

export interface GetSearchHistoryResponse {
  items: SearchHistoryItem[]
  total: number
  page: number
  page_size: number
}

export const semanticSearch = (data: SemanticSearchRequest) => {
  return apiPost<SemanticSearchResponse>('/search/semantic', data)
}

export const getSearchHistory = (page: number, page_size: number, search_type: string, target_type: string) => {
  return apiGet<GetSearchHistoryResponse>('/search/history', { params: { page, page_size, search_type, target_type } })
}