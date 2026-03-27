import { apiGet } from '@/utils/api'

export interface StatsOverview {
  total_notes: number
  total_knowledge_point: number
  total_categories: number
  total_tags: number
}

export interface CategoryStats {
  category_id: number
  category_name: string
  note_count: number
  knowledge_count: number
}

export interface GetCategoriesStatsResp {
  categories: CategoryStats[]
}

export const getStatsOverview = () => {
  return apiGet<StatsOverview>('/stats/overview')
}

export const getCategoryStats = () => {
  return apiGet<GetCategoriesStatsResp>('/stats/categories')
}
