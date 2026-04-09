import { useEffect, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  getKnowledgePoints,
  getKnowledgePoint,
  deleteKnowledgePoint,
  type KnowledgePoint
} from "@/services/knowledgeService"

function responseToPoint(res: Awaited<ReturnType<typeof getKnowledgePoint>>): KnowledgePoint {
  return {
    id: res.id,
    name: res.name,
    description: res.description,
    source_note_id: res.source_note_id,
    source_note_title: res.source_note_title,
    importance_score: res.importance_score,
    created_at: res.created_at,
  }
}

export default function KnowledgePage() {
  const { id: idParam } = useParams<{ id?: string }>()
  const navigate = useNavigate()
  const [knowledgePoints, setKnowledgePoints] = useState<KnowledgePoint[]>([])
  const [loading, setLoading] = useState(true)
  const [pointsLoaded, setPointsLoaded] = useState(false)
  const [selectedPoint, setSelectedPoint] = useState<KnowledgePoint | null>(null)

  useEffect(() => {
    setLoading(true)
    getKnowledgePoints({ page_size: 200 })
      .then(res => {
        setKnowledgePoints(res.knowledge_points || [])
      })
      .catch(err => console.error("Failed to fetch knowledge points", err))
      .finally(() => {
        setLoading(false)
        setPointsLoaded(true)
      })
  }, [])

  // 路由驱动：/knowledge 无选中；/knowledge/:id 展示对应知识点（列表无则单条拉取）
  useEffect(() => {
    let cancelled = false
    if (!idParam) {
      setSelectedPoint(null)
      return
    }
    const pid = Number(idParam)
    if (!Number.isFinite(pid) || pid < 1) {
      navigate("/knowledge", { replace: true })
      return
    }
    const fromList = knowledgePoints.find(p => p.id === pid)
    if (fromList) {
      setSelectedPoint(fromList)
      return
    }
    if (!pointsLoaded) return
    ;(async () => {
      try {
        const res = await getKnowledgePoint(pid)
        if (cancelled) return
        const kp = responseToPoint(res)
        setSelectedPoint(kp)
        setKnowledgePoints(prev => (prev.some(p => p.id === kp.id) ? prev : [kp, ...prev]))
      } catch {
        if (!cancelled) navigate("/knowledge", { replace: true })
      }
    })()
    return () => {
      cancelled = true
    }
  }, [idParam, knowledgePoints, pointsLoaded, navigate])

  // /knowledge 且无 id 时，默认打开列表第一项（保持可分享深链）
  useEffect(() => {
    if (idParam) return
    if (!pointsLoaded || knowledgePoints.length === 0) return
    navigate(`/knowledge/${knowledgePoints[0].id}`, { replace: true })
  }, [idParam, pointsLoaded, knowledgePoints, navigate])

  const handleDeletePoint = async () => {
    if (!selectedPoint) return
    const deletedId = selectedPoint.id
    try {
      await deleteKnowledgePoint(deletedId)
      setKnowledgePoints(prev => prev.filter(p => p.id !== deletedId))
      navigate("/knowledge", { replace: true })
    } catch (err) {
      console.error("Failed to delete knowledge point", err)
    }
  }

  const selectedId = idParam && /^\d+$/.test(idParam) ? Number(idParam) : null

  return (
    <div className="h-screen overflow-hidden p-6 box-border">
      <div className="max-w-6xl mx-auto h-full flex flex-col">
        <div className="flex justify-between items-center mb-6 shrink-0">
          <h1 className="text-3xl font-bold">知识点库</h1>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1 min-h-0">
          {/* 左侧列表 */}
          <div className="lg:col-span-1 min-h-0 overflow-hidden">
            <Card className="h-full flex flex-col overflow-hidden">
              <CardHeader className="shrink-0">
                <CardTitle className="text-lg">知识点列表</CardTitle>
              </CardHeader>
              <CardContent className="flex-1 overflow-y-auto">
                {loading ? (
                  <p className="text-muted-foreground">加载中...</p>
                ) : knowledgePoints.length === 0 ? (
                  <p className="text-muted-foreground">暂无知识点</p>
                ) : (
                  <ul className="space-y-2">
                    {knowledgePoints.map(point => (
                      <li
                        key={point.id}
                        className={`p-2 border rounded cursor-pointer transition-all text-sm ${selectedId === point.id
                            ? "bg-primary text-primary-foreground border-primary"
                            : "hover:bg-accent"
                          }`}
                        onClick={() => navigate(`/knowledge/${point.id}`)}
                      >
                        <p className="font-semibold truncate line-clamp-2">{point.name}</p>
                        <p className="text-xs opacity-70">
                          来源:{" "}
                          <Link
                            to={`/note/${point.source_note_id}`}
                            className="underline-offset-2 hover:underline"
                            onClick={e => e.stopPropagation()}
                          >
                            {point.source_note_title || `笔记 #${point.source_note_id}`}
                          </Link>
                        </p>
                      </li>
                    ))}
                  </ul>
                )}
              </CardContent>
            </Card>
          </div>

          {/* 右侧详情 */}
          <div className="lg:col-span-2">
            {selectedPoint ? (
              <Card className="h-full flex flex-col">
                <CardHeader>
                  <CardTitle className="text-base">
                    知识点详情
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 flex flex-col">
                  <div className="mb-4">
                    <h3 className="text-lg font-semibold mb-2">{selectedPoint.name}</h3>
                    <p className="text-sm text-muted-foreground mb-1">
                      来源笔记:{" "}
                      <Link
                        to={`/note/${selectedPoint.source_note_id}`}
                        className="text-blue-500 underline-offset-2 hover:underline"
                      >
                        {selectedPoint.source_note_title || `笔记 #${selectedPoint.source_note_id}`}
                      </Link>
                    </p>
                    <p className="text-xs text-muted-foreground">
                      创建于 {new Date(selectedPoint.created_at).toLocaleString()}
                    </p>
                  </div>
                  <p className="flex-1 whitespace-pre-wrap text-sm mb-4">
                    {selectedPoint.description}
                  </p>
                  <div className="flex gap-2">
                    <Button
                      size="lg"
                      variant="destructive"
                      onClick={handleDeletePoint}
                      className="ml-auto"
                    >
                      删除
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ) : (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">
                    未选择知识点
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-6">
                  <p className="text-muted-foreground text-center">选择一个知识点查看详情</p>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
