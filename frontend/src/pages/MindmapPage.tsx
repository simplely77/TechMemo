import { useEffect, useState, useRef } from "react"
import { useNavigate, useParams, useSearchParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getNotes, getNote, type Note } from "@/services/noteService"
import { getMindMap, getGlobalMindMap, generateGlobalMindMap, type MindMapNode, type GlobalMindMapNode } from "@/services/aiService"
import Tree from 'react-d3-tree'

export default function MindmapPage() {
  const { scope } = useParams<{ scope?: string }>()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const isGlobalView = scope === "global"
  const noteIdFromRoute =
    scope && scope !== "global" && /^\d+$/.test(scope) ? Number(scope) : null
  const kpRaw = searchParams.get("kp")
  const selectedKpId =
    kpRaw != null && /^\d+$/.test(kpRaw) ? Number(kpRaw) : null

  const [notes, setNotes] = useState<Note[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedNote, setSelectedNote] = useState<Note | null>(null)
  const [mindmapNodes, setMindmapNodes] = useState<MindMapNode[]>([])
  const [globalNodes, setGlobalNodes] = useState<GlobalMindMapNode[]>([])
  const [generating, setGenerating] = useState(false)
  const [hoveredNode, setHoveredNode] = useState<any>(null)
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 })
  const containerRef = useRef<HTMLDivElement>(null)
  const [translate, setTranslate] = useState({ x: 200, y: 200 })

  useEffect(() => {
    loadNotes()
  }, [])

  useEffect(() => {
    if (containerRef.current && mindmapNodes.length > 0) {
      const { offsetWidth,offsetHeight } = containerRef.current
      setTranslate({ x: offsetWidth / 8, y: offsetHeight / 2 }) // 根节点水平居中
    }
  }, [containerRef.current, mindmapNodes])

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

  // 按路由 scope 加载：无 scope | global | 笔记 id
  useEffect(() => {
    let cancelled = false
    if (!scope) {
      setMindmapNodes([])
      setGlobalNodes([])
      setSelectedNote(null)
      setHoveredNode(null)
      return
    }
    if (scope === "global") {
      setMindmapNodes([])
      setSelectedNote(null)
      setHoveredNode(null)
      ;(async () => {
        try {
          const res = await getGlobalMindMap()
          if (!cancelled) setGlobalNodes(res.nodes || [])
        } catch (err) {
          console.error("Failed to load global mindmap", err)
          if (!cancelled) setGlobalNodes([])
        }
      })()
      return () => {
        cancelled = true
      }
    }
    const noteId = Number(scope)
    if (!Number.isFinite(noteId) || noteId < 1) {
      navigate("/mindmap", { replace: true })
      return
    }
    setGlobalNodes([])
    setHoveredNode(null)
    ;(async () => {
      try {
        const [mapRes, noteFull] = await Promise.all([
          getMindMap(noteId),
          getNote(noteId).catch(() => null),
        ])
        if (cancelled) return
        setMindmapNodes(mapRes.nodes || [])
        setSelectedNote(noteFull)
      } catch (err) {
        console.error("Failed to load mindmap", err)
        if (!cancelled) {
          setMindmapNodes([])
          setSelectedNote(null)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [scope, navigate])

  // /mindmap 且无 scope 时，默认打开笔记列表第一篇（不含 /mindmap/global）
  useEffect(() => {
    if (scope) return
    if (loading || notes.length === 0) return
    navigate(`/mindmap/${notes[0].id}`, { replace: true })
  }, [scope, loading, notes, navigate])

  // URL 中的知识点 id 若不在当前树里（换笔记或数据变更），去掉无效 kp
  useEffect(() => {
    if (noteIdFromRoute == null || selectedKpId == null || mindmapNodes.length === 0) return
    if (!findNodeById(mindmapNodes, selectedKpId)) {
      navigate(`/mindmap/${noteIdFromRoute}`, { replace: true })
    }
  }, [noteIdFromRoute, selectedKpId, mindmapNodes, navigate])

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

  const getNodeColor = (score: number) => {
    if (score >= 0.7) return '#ef4444'
    if (score >= 0.4) return '#eab308'
    return '#22c55e'
  }

  const renderMindmapTree = (nodes: MindMapNode[], activeNoteId: number) => {
    if (!nodes || nodes.length === 0) return null

    const treeData = nodes[0]

    return (
      <div ref={containerRef} className="h-[calc(100vh-16rem)] w-full border rounded bg-background relative overflow-hidden">
        <Tree
          data={treeData}
          orientation="horizontal"
          nodeSize={{ x: 220, y: 120 }}
          translate={translate}
          pathFunc="diagonal"
          zoom={1}
          scaleExtent={{ min: 0.5, max: 1.5 }}
          enableLegacyTransitions
          pathClassFunc={()=>"custom-link"}

          renderCustomNodeElement={({ nodeDatum }) => {
            const score = (nodeDatum as any).importance_score || 0
            const color = getNodeColor(score)
            const nodeData = nodeDatum as unknown as MindMapNode
            const name = nodeDatum.name
            const lines = name.match(/.{1,8}/g) || []
            const isKpSelected = selectedKpId != null && nodeData.id === selectedKpId

            return (
              <g
                onMouseEnter={(e) => {
                  setHoveredNode(nodeData)
                  setMousePos({ x: e.clientX, y: e.clientY })
                }}
                onMouseLeave={() => setHoveredNode(null)}
                onClick={(e) => {
                  e.stopPropagation()
                  navigate(`/mindmap/${activeNoteId}?kp=${nodeData.id}`, { replace: true })
                }}
                style={{ cursor: 'pointer' }}
              >
                <rect
                  width="160"
                  height="50"
                  x="-80"
                  y="-25"
                  rx="12"
                  fill={color}
                  stroke={isKpSelected ? '#fafafa' : 'none'}
                  strokeWidth={isKpSelected ? 3 : 0}
                />
                <text
                  fill="white"
                  textAnchor="middle"
                  stroke="none"
                  strokeWidth="0"
                  fontSize="12"
                >
                  {lines.map((line, i) => (
                    <tspan key={i} x="0" dy={i === 0 ? 0 : 14}>
                      {line}
                    </tspan>
                  ))}
                </text>
              </g>
            )
          }}
        />

        {/* 悬浮弹窗 */}
        {hoveredNode && (
          <div
            className="fixed z-50 bg-popover text-popover-foreground rounded-lg shadow-lg border p-4 max-w-sm"
            style={{
              left: mousePos.x + 20,
              top: mousePos.y - 50,
              pointerEvents: 'none'
            }}
          >
            <div className="space-y-2">
              {typeof hoveredNode.id === 'number' && (
                <div>
                  <p className="text-xs text-muted-foreground">知识点 ID</p>
                  <p className="font-mono text-sm">{hoveredNode.id}</p>
                </div>
              )}
              <div>
                <p className="text-xs text-muted-foreground">名称</p>
                <p className="font-semibold">{hoveredNode.name}</p>
              </div>
              {hoveredNode.description && (
                <div>
                  <p className="text-xs text-muted-foreground">描述</p>
                  <p className="text-sm">{hoveredNode.description}</p>
                </div>
              )}
              <div>
                <p className="text-xs text-muted-foreground mb-1">重要度</p>
                <div className="flex items-center gap-2">
                  <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                    <div
                      className="h-full transition-all"
                      style={{
                        width: `${(hoveredNode.importance_score || 0) * 100}%`,
                        backgroundColor: getNodeColor(hoveredNode.importance_score || 0)
                      }}
                    />
                  </div>
                  <span className="text-xs font-medium">
                    {((hoveredNode.importance_score || 0) * 100).toFixed(0)}%
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    )
  }


  return (
    <div className="h-screen overflow-hidden p-6 box-border">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">思维导图</h1>
          <div className="flex gap-2">
            <Button onClick={() => navigate("/mindmap/global")} variant="outline">
              查看全局知识图谱
            </Button>
            <Button onClick={handleGenerateGlobalMindmap} disabled={generating}>
              {generating ? "生成中..." : "生成全局图谱"}
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* 左侧笔记列表 */}
          <div className="lg:col-span-3">
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
                  <div className="space-y-1 max-h-[calc(100vh-20rem)] overflow-y-auto">
                    {notes.map((note) => (
                      <div
                        key={note.id}
                        className={`p-2 rounded cursor-pointer text-sm transition-all ${noteIdFromRoute === note.id
                          ? "bg-primary text-primary-foreground shadow-sm"
                          : "hover:bg-accent"
                          }`}
                        onClick={() => navigate(`/mindmap/${note.id}`)}
                      >
                        <p className="font-semibold truncate line-clamp-2">{note.title}</p>
                        <p className="text-xs opacity-70">{note.category.name}</p>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* 思维导图展示 */}
          <div className="lg:col-span-9">
            <Card className="h-full overflow-hidden">
              <CardHeader>
                <CardTitle className="text-lg">
                  {isGlobalView
                    ? "全局知识图谱"
                    : noteIdFromRoute != null
                      ? `${selectedNote?.title ?? `笔记 #${noteIdFromRoute}`} - 思维导图`
                      : "选择笔记查看思维导图"}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {noteIdFromRoute != null ? (
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
                    <div>
                      {renderMindmapTree(mindmapNodes, noteIdFromRoute)}
                    </div>
                  )
                ) : isGlobalView ? (
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
                    <div className="max-h-[calc(100vh-16rem)] overflow-y-auto">
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {globalNodes.map((node) => (
                          <Card
                            key={node.id}
                            role="button"
                            tabIndex={0}
                            onKeyDown={(e) => {
                              if (e.key === "Enter" || e.key === " ") {
                                e.preventDefault()
                                navigate(`/mindmap/${node.note_id}`)
                              }
                            }}
                            onClick={() => navigate(`/mindmap/${node.note_id}`)}
                            className="hover:shadow-lg transition-all cursor-pointer border-l-4"
                            style={{
                              borderLeftColor: `hsl(var(--primary) / ${node.importance_score})`
                            }}
                          >
                            <CardHeader className="pb-3">
                              <div className="flex items-start justify-between gap-2">
                                <CardTitle className="text-sm font-semibold">{node.name}</CardTitle>
                                <div className={`shrink-0 px-2 py-0.5 rounded-full text-xs font-medium ${node.importance_score >= 0.7
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
        </div>
      </div>
    </div>
  )
}
