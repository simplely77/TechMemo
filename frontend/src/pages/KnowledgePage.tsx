import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  getKnowledgePoints,
  updateKnowledgePoint,
  deleteKnowledgePoint,
  type KnowledgePoint
} from "@/services/knowledgeService"

export default function KnowledgePage() {
  const [knowledgePoints, setKnowledgePoints] = useState<KnowledgePoint[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedPoint, setSelectedPoint] = useState<KnowledgePoint | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [editContent, setEditContent] = useState("")
  const [editCategory, setEditCategory] = useState("")

  useEffect(() => {
    setLoading(true)
    getKnowledgePoints()
      .then(res => {
        const points = res.knowledge_points || []
        setKnowledgePoints(points)
      })
      .catch(err => console.error("Failed to fetch knowledge points", err))
      .finally(() => setLoading(false))
  }, [])

  const handleSelectPoint = (point: KnowledgePoint) => {
    setSelectedPoint(point)
    setEditContent(point.description)
    setEditCategory(point.name)
    setIsEditing(false)
  }

  const handleUpdatePoint = async () => {
    if (!selectedPoint) return
    try {
      const updated = await updateKnowledgePoint(selectedPoint.id, {
        name: editCategory,
        description: editContent
      })
      setKnowledgePoints(
        knowledgePoints.map(p => p.id === selectedPoint.id ? updated : p)
      )
      setSelectedPoint(updated)
      setIsEditing(false)
    } catch (err) {
      console.error("Failed to update knowledge point", err)
    }
  }

  const handleDeletePoint = async () => {
    if (!selectedPoint) return
    try {
      await deleteKnowledgePoint(selectedPoint.id)
      setKnowledgePoints(knowledgePoints.filter(p => p.id !== selectedPoint.id))
      setSelectedPoint(null)
    } catch (err) {
      console.error("Failed to delete knowledge point", err)
    }
  }

  return (
    <div className="h-screen overflow-hidden p-6 box-border">
      <div className="max-w-6xl mx-auto h-full flex flex-col">
        <div className="flex justify-between items-center mb-6 flex-shrink-0">
          <h1 className="text-3xl font-bold">知识点库</h1>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1 min-h-0">
          {/* 左侧列表 */}
          <div className="lg:col-span-1 min-h-0 overflow-hidden">
            <Card className="h-full flex flex-col overflow-hidden">
              <CardHeader className="flex-shrink-0">
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
                        className={`p-2 border rounded cursor-pointer transition-all text-sm ${
                          selectedPoint?.id === point.id
                            ? "bg-primary text-primary-foreground border-primary"
                            : "hover:bg-accent"
                        }`}
                        onClick={() => handleSelectPoint(point)}
                      >
                        <p className="font-semibold truncate line-clamp-2">{point.name}</p>
                        {point.source_note_title && (
                          <p className="text-xs opacity-70">来源: {point.source_note_title}</p>
                        )}
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
                        {selectedPoint.source_note_title && (
                          <p className="text-sm text-muted-foreground mb-1">
                            来源笔记: {selectedPoint.source_note_title}
                          </p>
                        )}
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
