import request from '../utils/request'

export interface StatsOverview {
  total_notes: number
  total_knowledge_points: number
  total_categories: number
  total_tags: number
}

export interface CategoryStats {
  category_id: number
  category_name: string
  note_count: number
}

export const getStatsOverview = () => {
  return request.get<{ code: number; message: string; data: StatsOverview }>('/stats/overview')
}

export const getCategoryStats = () => {
  return request.get<{ code: number; message: string; data: CategoryStats[] }>('/stats/categories')
}
