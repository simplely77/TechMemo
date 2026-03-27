import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getStatsOverview, getCategoryStats, type StatsOverview, type GetCategoriesStatsResp } from "@/services/statsService"

export default function StatsPage() {
  const [stats, setStats] = useState<StatsOverview | null>(null)
  const [categoryStats, setCategoryStats] = useState<GetCategoriesStatsResp | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadStats()
  }, [])

  const loadStats = async () => {
    setLoading(true)
    try {
      const [overviewRes, categoryRes] = await Promise.all([
        getStatsOverview(),
        getCategoryStats()
      ])
      setStats(overviewRes)
      setCategoryStats(categoryRes)
    } catch (err) {
      console.error("Failed to load stats", err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen p-6">
        <p className="text-muted-foreground">加载中...</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">统计分析</h1>

        {/* 关键指标 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">笔记总数</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{stats?.total_notes || 0}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">知识点</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{stats?.total_knowledge_point || 0}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">分类</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{stats?.total_categories || 0}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">标签</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{stats?.total_tags || 0}</p>
            </CardContent>
          </Card>
        </div>

        {/* 分类统计 */}
        <Card>
          <CardHeader>
            <CardTitle>分类统计</CardTitle>
          </CardHeader>
          <CardContent>
            {!categoryStats?.categories || categoryStats.categories.length === 0 ? (
              <p className="text-muted-foreground text-center">暂无分类数据</p>
            ) : (
              <div className="space-y-4">
                {categoryStats.categories.map((cat) => {
                  const maxNotes = Math.max(...categoryStats.categories.map(c => c.note_count), 1)
                  return (
                    <div key={cat.category_id} className="flex items-center justify-between">
                      <div className="flex-1">
                        <p className="text-sm font-medium">{cat.category_name}</p>
                        <p className="text-xs text-muted-foreground">知识点: {cat.knowledge_count}</p>
                      </div>
                      <div className="flex items-center gap-2 ml-4">
                        <div className="w-32 h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className="h-full bg-primary"
                            style={{
                              width: `${(cat.note_count / maxNotes) * 100}%`
                            }}
                          />
                        </div>
                        <span className="text-sm font-semibold w-8 text-right">{cat.note_count}</span>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
