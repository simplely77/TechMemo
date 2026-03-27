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
        const points = Array.isArray(res) ? res : []
        setKnowledgePoints(points)
      })
      .catch(err => console.error("Failed to fetch knowledge points", err))
      .finally(() => setLoading(false))
  }, [])

  const handleSelectPoint = (point: KnowledgePoint) => {
    setSelectedPoint(point)
    setEditContent(point.content)
    setEditCategory(point.category || "")
    setIsEditing(false)
  }

  const handleUpdatePoint = async () => {
    if (!selectedPoint) return
    try {
      const updated = await updateKnowledgePoint(selectedPoint.id, {
        content: editContent,
        category: editCategory
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
    <div className="min-h-screen p-6">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">知识点库</h1>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* 左侧列表 */}
          <div className="lg:col-span-1">
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">知识点列表</CardTitle>
              </CardHeader>
              <CardContent>
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
                        <p className="font-semibold truncate line-clamp-2">{point.content}</p>
                        {point.category && (
                          <p className="text-xs opacity-70">{point.category}</p>
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
                    {isEditing ? "编辑知识点" : "知识点详情"}
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 flex flex-col">
                  {isEditing ? (
                    <>
                      <input
                        type="text"
                        placeholder="分类"
                        value={editCategory}
                        onChange={(e) => setEditCategory(e.target.value)}
                        className="mb-3 px-3 py-2 border rounded"
                      />
                      <textarea
                        value={editContent}
                        onChange={(e) => setEditContent(e.target.value)}
                        className="flex-1 px-3 py-2 border rounded mb-3 resize-none"
                      />
                      <div className="flex gap-2">
                        <Button size="sm" onClick={handleUpdatePoint}>
                          保存
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setIsEditing(false)}
                        >
                          取消
                        </Button>
                      </div>
                    </>
                  ) : (
                    <>
                      {selectedPoint.category && (
                        <p className="text-sm text-muted-foreground mb-2">
                          分类: {selectedPoint.category}
                        </p>
                      )}
                      <p className="text-sm text-muted-foreground mb-4">
                        创建于 {new Date(selectedPoint.created_time).toLocaleString()}
                      </p>
                      <p className="flex-1 whitespace-pre-wrap text-sm mb-4">
                        {selectedPoint.content}
                      </p>
                      <div className="flex gap-2">
                        <Button size="sm" onClick={() => setIsEditing(true)}>
                          编辑
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={handleDeletePoint}
                        >
                          删除
                        </Button>
                      </div>
                    </>
                  )}
                </CardContent>
              </Card>
            ) : (
              <Card>
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
