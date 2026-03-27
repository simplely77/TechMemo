import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

interface Mindmap {
  id: number
  title: string
  description: string
  content: string
  created_at: string
}

export default function MindmapPage() {
  const [mindmaps, setMindmaps] = useState<Mindmap[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedMindmap, setSelectedMindmap] = useState<Mindmap | null>(null)

  useEffect(() => {
    // TODO: 调用 API 获取思维导图列表
    // 示例数据
    setLoading(false)
  }, [])

  const handleGenerateMindmap = async () => {
    // TODO: 调用 API 生成新思维导图
    console.log("Generate new mindmap")
  }

  const handleDeleteMindmap = async () => {
    if (!selectedMindmap) return
    // TODO: 调用 API 删除思维导图
    setMindmaps(mindmaps.filter(m => m.id !== selectedMindmap.id))
    setSelectedMindmap(null)
  }

  const handleDownloadMindmap = () => {
    if (!selectedMindmap) return
    // TODO: 调用 API 下载思维导图
    console.log("Download mindmap:", selectedMindmap.id)
  }

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">思维导图</h1>
          <Button onClick={handleGenerateMindmap}>生成新导图</Button>
        </div>

        {loading ? (
          <p className="text-muted-foreground">加载中...</p>
        ) : mindmaps.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-muted-foreground text-center">暂无思维导图，点击"生成新导图"创建</p>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
            {mindmaps.map((mindmap) => (
              <Card
                key={mindmap.id}
                className="cursor-pointer hover:shadow-lg transition-all"
                onClick={() => setSelectedMindmap(mindmap)}
              >
                <CardHeader>
                  <CardTitle className="text-lg">{mindmap.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground mb-4">{mindmap.description}</p>
                  <Button variant="outline" size="sm" className="w-full">查看</Button>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {selectedMindmap && (
          <div className="mt-8">
            <Card>
              <CardHeader>
                <CardTitle>{selectedMindmap.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="bg-muted p-4 rounded min-h-96 flex items-center justify-center mb-4">
                  <p className="text-muted-foreground">思维导图可视化区域</p>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={handleDownloadMindmap}>
                    下载
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={handleDeleteMindmap}
                  >
                    删除
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </div>
  )
}
