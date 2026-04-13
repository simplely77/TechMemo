import { useEffect, useState } from 'react'
import { useAuthStore } from '../store/authStore'
import { useNavigate } from 'react-router-dom'
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getStatsOverview, type StatsOverview } from '@/services/statsService'

export default function HomePage() {
  const { user, clearAuth } = useAuthStore()
  const navigate = useNavigate()
  const [stats, setStats] = useState<StatsOverview | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadStats()
  }, [])

  const loadStats = async () => {
    try {
      const res = await getStatsOverview()
      setStats(res)
    } catch (err) {
      console.error("Failed to load stats", err)
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    clearAuth()
    navigate('/login')
  }

  const features = [
    {
      title: '笔记管理',
      description: '创建、编辑和管理你的技术笔记',
      icon: '📝',
      action: () => navigate('/note')
    },
    {
      title: '知识点',
      description: 'AI 自动提取的知识点库',
      icon: '💡',
      action: () => navigate('/knowledge')
    },
    {
      title: '思维导图',
      description: '可视化你的知识结构',
      icon: '🗺️',
      action: () => navigate('/mindmap')
    },
    {
      title: '搜索',
      description: '智能搜索笔记和知识点',
      icon: '🔍',
      action: () => navigate('/search')
    },
    {
      title: '知识库问答',
      description: '基于你的知识库回答问题',
      icon: '💬',
      action: () => navigate('/qa')
    },
    {
      title: '统计分析',
      description: '查看学习进度和数据统计',
      icon: '📊',
      action: () => navigate('/stats')
    }
  ]

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="border-b border-zinc-800 bg-zinc-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🧠</span>
            <h1 className="text-xl font-bold">TechMemo</h1>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">Hi, {user?.username}</span>
            <Button variant="outline" size="sm" onClick={handleLogout}>
              退出
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-12">
        <div className="max-w-6xl mx-auto">
          {/* Welcome Section */}
          <div className="mb-12 text-center">
            <h2 className="text-4xl font-bold mb-3 bg-linear-to-r from-white to-zinc-400 bg-clip-text text-transparent">
              欢迎来到 TechMemo
            </h2>
            <p className="text-muted-foreground text-lg">
              你的个人技术知识库
            </p>
          </div>

          {/* Quick Stats - 放在功能卡片前面 */}
          <div className="mb-12 grid grid-cols-2 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardDescription className="text-xs">笔记总数</CardDescription>
                <CardTitle className="text-2xl">
                  {loading ? '--' : stats?.total_notes || 0}
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardDescription className="text-xs">知识点</CardDescription>
                <CardTitle className="text-2xl">
                  {loading ? '--' : stats?.total_knowledge_point || 0}
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardDescription className="text-xs">分类</CardDescription>
                <CardTitle className="text-2xl">
                  {loading ? '--' : stats?.total_categories || 0}
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardDescription className="text-xs">标签</CardDescription>
                <CardTitle className="text-2xl">
                  {loading ? '--' : stats?.total_tags || 0}
                </CardTitle>
              </CardHeader>
            </Card>
          </div>

          {/* Feature Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <Card
                key={index}
                className="cursor-pointer transition-all hover:shadow-lg hover:-translate-y-1 border-2 hover:border-primary/50"
                onClick={feature.action}
              >
                <CardHeader>
                  <div className="text-4xl mb-2">{feature.icon}</div>
                  <CardTitle className="text-xl">{feature.title}</CardTitle>
                  <CardDescription className="text-sm">
                    {feature.description}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button variant="ghost" size="sm" className="w-full">
                    进入 →
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </main>
    </div>
  )
}
