import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getNotes, type Note } from "@/services/noteService"
import { getMindMap, getGlobalMindMap, generateGlobalMindMap, type MindMapNode, type GlobalMindMapNode } from "@/services/aiService"

type ViewMode = "note" | "global"

export default function MindmapPage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedNote, setSelectedNote] = useState<Note | null>(null)
  const [mindmapNodes, setMindmapNodes] = useState<MindMapNode[]>([])
  const [globalNodes, setGlobalNodes] = useState<GlobalMindMapNode[]>([])
  const [viewMode, setViewMode] = useState<ViewMode>("note")
  const [generating, setGenerating] = useState(false)
  const [selectedNodeId, setSelectedNodeId] = useState<number | null>(null)

  useEffect(() => {
    loadNotes()
  }, [])

  const loadNotes = async () => {
    setLoading(true)
    try {
      const res = await getNotes({ page_size: 100 })
      setNotes(res.notes || [])
    } catch (err) {
      console.error("Failed to load notes", err)
    } finally {
      setLoading(false)
    }
  }

  const handleSelectNote = async (note: Note) => {
    setSelectedNote(note)
    setViewMode("note")
    setSelectedNodeId(null)
    try {
      const res = await getMindMap(note.id)
      console.log("思维导图数据:", res)
      setMindmapNodes(res.nodes || [])
    } catch (err) {
      console.error("Failed to load mindmap", err)
      setMindmapNodes([])
    }
  }

  const handleLoadGlobalMindmap = async () => {
    setViewMode("global")
    setSelectedNote(null)
    setSelectedNodeId(null)
    try {
      const res = await getGlobalMindMap()
      console.log("全局思维导图数据:", res)
      setGlobalNodes(res.nodes || [])
    } catch (err) {
      console.error("Failed to load global mindmap", err)
      setGlobalNodes([])
    }
  }

  const handleGenerateGlobalMindmap = async () => {
    setGenerating(true)
    try {
      await generateGlobalMindMap()
      alert("全局思维导图生成任务已启动，请稍后刷新查看")
    } catch (err) {
      console.error("Failed to generate global mindmap", err)
      alert("生成失败")
    } finally {
      setGenerating(false)
    }
  }

  const findNodeById = (nodes: MindMapNode[], id: number): MindMapNode | null => {
    for (const node of nodes) {
      if (node.id === id) return node
      if (node.children && node.children.length > 0) {
        const found = findNodeById(node.children, id)
        if (found) return found
      }
    }
    return null
  }

  const getSelectedNodeDetails = () => {
    if (!selectedNodeId) return null
    return findNodeById(mindmapNodes, selectedNodeId)
  }

  const renderMindmapTree = (nodes: MindMapNode[], level = 0) => {
    if (!nodes || nodes.length === 0) return null

    return (
      <div className={level > 0 ? "ml-8 mt-3" : ""}>
        {nodes.map((node, index) => (
          <div key={node.id} className="relative">
            {/* 连接线 */}
            {level > 0 && (
              <>
                <div className="absolute -left-8 top-5 w-6 h-0.5 bg-gradient-to-r from-primary/30 to-primary/60" />
                {index < nodes.length - 1 && (
                  <div className="absolute -left-8 top-5 w-0.5 bg-primary/30" style={{ height: 'calc(100% + 12px)' }} />
                )}
              </>
            )}

            {/* 节点卡片 */}
            <div
              className={`group relative mb-3 cursor-pointer transition-all duration-200 ${
                selectedNodeId === node.id ? 'scale-105' : 'hover:scale-102'
              }`}
              onClick={(e) => {
                e.stopPropagation()
                setSelectedNodeId(node.id)
              }}
            >
              <div
                className={`relative rounded-lg p-3 shadow-sm border-2 transition-all ${
                  selectedNodeId === node.id
                    ? 'border-primary bg-primary/5 shadow-md'
                    : 'border-border bg-card hover:border-primary/50 hover:shadow-md'
                }`}
                style={{
                  background: selectedNodeId === node.id
                    ? 'linear-gradient(135deg, hsl(var(--primary) / 0.05) 0%, hsl(var(--primary) / 0.02) 100%)'
                    : undefined
                }}
              >
                {/* 重要度指示器 */}
                <div
                  className="absolute top-0 left-0 h-full w-1 rounded-l-lg"
                  style={{
                    background: `linear-gradient(to bottom, hsl(var(--primary) / ${node.importance_score}), hsl(var(--primary) / ${node.importance_score * 0.5}))`
                  }}
                />

                <div className="flex items-center gap-2 pl-2">
                  {/* 节点图标 */}
                  <div
                    className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold transition-colors ${
                      selectedNodeId === node.id
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-primary/10 text-primary group-hover:bg-primary/20'
                    }`}
                  >
                    {level === 0 ? '🎯' : level === 1 ? '📌' : '•'}
                  </div>

                  {/* 节点名称 */}
                  <div className="flex-1 min-w-0">
                    <p className={`font-medium text-sm truncate ${
                      selectedNodeId === node.id ? 'text-primary' : ''
                    }`}>
                      {node.name}
                    </p>
                    {node.children && node.children.length > 0 && (
                      <p className="text-xs text-muted-foreground">
                        {node.children.length} 个子节点
                      </p>
                    )}
                  </div>

                  {/* 重要度标签 */}
                  <div className={`flex-shrink-0 px-2 py-0.5 rounded-full text-xs font-medium ${
                    node.importance_score >= 0.7
                      ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                      : node.importance_score >= 0.4
                      ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                      : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                  }`}>
                    {(node.importance_score * 100).toFixed(0)}%
                  </div>
                </div>
              </div>
            </div>

            {/* 子节点 */}
            {node.children && node.children.length > 0 && (
              <div className="relative">
                {renderMindmapTree(node.children, level + 1)}
              </div>
            )}
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">思维导图</h1>
          <div className="flex gap-2">
            <Button onClick={handleLoadGlobalMindmap} variant="outline">
              查看全局知识图谱
            </Button>
            <Button onClick={handleGenerateGlobalMindmap} disabled={generating}>
              {generating ? "生成中..." : "生成全局图谱"}
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* 左侧笔记列表 */}
          <div className="lg:col-span-2">
            <Card className="sticky top-6">
              <CardHeader>
                <CardTitle className="text-base">笔记列表</CardTitle>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <p className="text-muted-foreground text-sm">加载中...</p>
                ) : notes.length === 0 ? (
                  <p className="text-muted-foreground text-sm">暂无笔记</p>
                ) : (
                  <div className="space-y-1 max-h-[600px] overflow-y-auto">
                    {notes.map((note) => (
                      <div
                        key={note.id}
                        className={`p-2 rounded cursor-pointer text-xs transition-all ${
                          selectedNote?.id === note.id
                            ? "bg-primary text-primary-foreground shadow-sm"
                            : "hover:bg-accent"
                        }`}
                        onClick={() => handleSelectNote(note)}
                      >
                        <p className="font-semibold truncate">{note.title}</p>
                        <p className="text-[10px] opacity-70 truncate">{note.category.name}</p>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* 中间思维导图展示 */}
          <div className={selectedNodeId ? "lg:col-span-7" : "lg:col-span-10"}>
            <Card className="h-full">
              <CardHeader>
                <CardTitle className="text-lg">
                  {viewMode === "note" && selectedNote
                    ? `${selectedNote.title} - 思维导图`
                    : viewMode === "global"
                    ? "全局知识图谱"
                    : "选择笔记查看思维导图"}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {viewMode === "note" && selectedNote ? (
                  mindmapNodes.length === 0 ? (
                    <div className="text-center py-20">
                      <div className="text-6xl mb-4">🤖</div>
                      <p className="text-muted-foreground mb-2 font-medium">
                        该笔记还没有生成思维导图
                      </p>
                      <p className="text-sm text-muted-foreground">
                        请先在笔记页面点击"AI 分析"按钮处理笔记
                      </p>
                    </div>
                  ) : (
                    <div className="max-h-[700px] overflow-y-auto pr-2">
                      {renderMindmapTree(mindmapNodes)}
                    </div>
                  )
                ) : viewMode === "global" ? (
                  globalNodes.length === 0 ? (
                    <div className="text-center py-20">
                      <div className="text-6xl mb-4">🌐</div>
                      <p className="text-muted-foreground mb-2 font-medium">
                        还没有生成全局知识图谱
                      </p>
                      <p className="text-sm text-muted-foreground">
                        点击"生成全局图谱"按钮开始生成
                      </p>
                    </div>
                  ) : (
                    <div className="max-h-[700px] overflow-y-auto">
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {globalNodes.map((node) => (
                          <Card
                            key={node.id}
                            className="hover:shadow-lg transition-all cursor-pointer border-l-4"
                            style={{
                              borderLeftColor: `hsl(var(--primary) / ${node.importance_score})`
                            }}
                          >
                            <CardHeader className="pb-3">
                              <div className="flex items-start justify-between gap-2">
                                <CardTitle className="text-sm font-semibold">{node.name}</CardTitle>
                                <div className={`flex-shrink-0 px-2 py-0.5 rounded-full text-xs font-medium ${
                                  node.importance_score >= 0.7
                                    ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                                    : node.importance_score >= 0.4
                                    ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                                    : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                }`}>
                                  {(node.importance_score * 100).toFixed(0)}%
                                </div>
                              </div>
                            </CardHeader>
                            <CardContent className="pt-0">
                              <p className="text-xs text-muted-foreground line-clamp-2 mb-2">
                                {node.description}
                              </p>
                              <p className="text-xs text-muted-foreground">
                                来源笔记 ID: {node.note_id}
                              </p>
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    </div>
                  )
                ) : (
                  <div className="text-center py-20">
                    <div className="text-6xl mb-4">🗺️</div>
                    <p className="text-muted-foreground">
                      从左侧选择一篇笔记查看其思维导图
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* 右侧详情面板 */}
          {selectedNodeId && viewMode === "note" && (
            <div className="lg:col-span-3">
              <Card className="sticky top-6">
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-base">节点详情</CardTitle>
                    <button
                      onClick={() => setSelectedNodeId(null)}
                      className="text-muted-foreground hover:text-foreground transition-colors"
                    >
                      ✕
                    </button>
                  </div>
                </CardHeader>
                <CardContent>
                  {(() => {
                    const node = getSelectedNodeDetails()
                    if (!node) return <p className="text-sm text-muted-foreground">未找到节点</p>

                    return (
                      <div className="space-y-4">
                        {/* 节点名称 */}
                        <div>
                          <p className="text-xs text-muted-foreground mb-1">名称</p>
                          <p className="font-semibold text-sm">{node.name}</p>
                        </div>

                        {/* 描述 */}
                        <div>
                          <p className="text-xs text-muted-foreground mb-1">描述</p>
                          <p className="text-sm leading-relaxed">{node.description}</p>
                        </div>

                        {/* 重要度 */}
                        <div>
                          <p className="text-xs text-muted-foreground mb-2">重要度</p>
                          <div className="space-y-2">
                            <div className="flex items-center gap-2">
                              <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                                <div
                                  className="h-full bg-gradient-to-r from-primary to-primary/60 transition-all"
                                  style={{ width: `${node.importance_score * 100}%` }}
                                />
                              </div>
                              <span className="text-xs font-medium">
                                {(node.importance_score * 100).toFixed(0)}%
                              </span>
                            </div>
                            <p className="text-xs text-muted-foreground">
                              {node.importance_score >= 0.7
                                ? "🔴 高重要度"
                                : node.importance_score >= 0.4
                                ? "🟡 中等重要度"
                                : "🟢 低重要度"}
                            </p>
                          </div>
                        </div>

                        {/* 子节点统计 */}
                        {node.children && node.children.length > 0 && (
                          <div>
                            <p className="text-xs text-muted-foreground mb-1">子节点</p>
                            <div className="flex items-center gap-2">
                              <div className="flex-1 bg-accent rounded p-2 text-center">
                                <p className="text-lg font-bold text-primary">{node.children.length}</p>
                                <p className="text-xs text-muted-foreground">个子节点</p>
                              </div>
                            </div>
                          </div>
                        )}

                        {/* 节点 ID */}
                        <div>
                          <p className="text-xs text-muted-foreground mb-1">节点 ID</p>
                          <p className="text-xs font-mono bg-muted px-2 py-1 rounded">
                            {node.id}
                          </p>
                        </div>
                      </div>
                    )
                  })()}
                </CardContent>
              </Card>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
