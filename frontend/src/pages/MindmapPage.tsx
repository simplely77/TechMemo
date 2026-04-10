import { useEffect, useState, useRef, useMemo, useCallback } from "react"
import { useNavigate, useParams, useSearchParams } from "react-router-dom"
import ForceGraph2D, { type ForceGraphMethods } from "react-force-graph-2d"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getNotes, getNote, type Note } from "@/services/noteService"
import {
  getMindMap,
  getGlobalMindMap,
  generateGlobalMindMap,
  getTaskStatus,
  type MindMapNode,
  type GlobalMindMapNode,
  type GlobalMindMapEdge,
} from "@/services/aiService"
import Tree from "react-d3-tree"
import { Hand, MousePointerClick, ZoomIn } from "lucide-react"

/** 全局力导向图节点画布尺寸（与 pointer 命中区共用） */
function measureGlobalNodeLabel(
  node: { name?: string; importance_score?: number; x?: number; y?: number },
  ctx: CanvasRenderingContext2D,
  globalScale: number,
) {
  const label = String(node.name || "")
  const fontSize = 11 / globalScale
  const padX = 10 / globalScale
  const padY = 6 / globalScale
  const maxPx = 200 / globalScale
  ctx.font = `500 ${fontSize}px ui-sans-serif, system-ui, sans-serif`
  let display = label
  while (display.length > 1 && ctx.measureText(display + "…").width > maxPx) {
    display = display.slice(0, -1)
  }
  if (display !== label && label.length > 0) display += "…"
  const tw = Math.min(ctx.measureText(display).width + padX * 2, maxPx + padX * 2)
  const th = fontSize + padY * 2
  return { display, tw, th, fontSize, padX, padY, x: node.x as number, y: node.y as number }
}

function paintGlobalMindMapNode(
  node: { name?: string; importance_score?: number; x?: number; y?: number },
  ctx: CanvasRenderingContext2D,
  globalScale: number,
  getFill: (score: number) => string,
  hitColor?: string,
) {
  const m = measureGlobalNodeLabel(node, ctx, globalScale)
  const { x, y } = m
  if (x == null || y == null || Number.isNaN(x) || Number.isNaN(y)) return

  const rx = 9 / globalScale
  const left = x - m.tw / 2
  const top = y - m.th / 2
  const score = (node.importance_score ?? 0) as number

  ctx.save()
  if (!hitColor) {
    ctx.shadowColor = "rgba(0, 0, 0, 0.14)"
    ctx.shadowBlur = 10 / globalScale
    ctx.shadowOffsetY = 3 / globalScale
  }

  ctx.beginPath()
  if (typeof ctx.roundRect === "function") {
    ctx.roundRect(left, top, m.tw, m.th, rx)
  } else {
    ctx.rect(left, top, m.tw, m.th)
  }
  ctx.fillStyle = hitColor ?? getFill(score)
  ctx.fill()

  if (!hitColor) {
    ctx.shadowColor = "transparent"
    ctx.shadowBlur = 0
    ctx.shadowOffsetY = 0
    ctx.strokeStyle = "rgba(255, 255, 255, 0.22)"
    ctx.lineWidth = 1 / globalScale
    ctx.stroke()

    ctx.fillStyle = "rgba(255, 255, 255, 0.97)"
    ctx.textAlign = "center"
    ctx.textBaseline = "middle"
    ctx.fillText(m.display, x, y)
  }
  ctx.restore()
}

/** 与画布 pill 标签大致一致，用于在图坐标里估算「从节点中心到块边缘」距离，避免箭头被节点盖住 */
function estimateGlobalNodeAnchorReach(name: string): number {
  const label = String(name || "·")
  const pad = 10
  const maxW = 200
  const charW = 6.2
  const halfW = Math.min(label.length * charW + pad, maxW) / 2
  const halfH = (11 + 12) / 2
  return Math.hypot(halfW, halfH) + 14
}

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
  const [globalEdges, setGlobalEdges] = useState<GlobalMindMapEdge[]>([])
  const [generating, setGenerating] = useState(false)
  /** 全局图谱异步任务 task_id，非空时由 effect 轮询状态 */
  const [globalGenPollTaskId, setGlobalGenPollTaskId] = useState<string | null>(null)
  const [hoveredNode, setHoveredNode] = useState<any>(null)
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 })
  const containerRef = useRef<HTMLDivElement>(null)
  const [translate, setTranslate] = useState({ x: 200, y: 200 })
  const globalFgRef = useRef<ForceGraphMethods | undefined>(undefined)

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
      setGlobalEdges([])
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
          if (!cancelled) {
            setGlobalNodes(res.nodes || [])
            setGlobalEdges(res.edges || [])
          }
        } catch (err) {
          console.error("Failed to load global mindmap", err)
          if (!cancelled) {
            setGlobalNodes([])
            setGlobalEdges([])
          }
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
    setGlobalEdges([])
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

  // 全局思维导图任务：轮询直到 completed / failed，与笔记页 AI 状态轮询类似
  useEffect(() => {
    if (!globalGenPollTaskId) return
    let cancelled = false
    let intervalId: ReturnType<typeof setInterval> | undefined

    const stop = () => {
      if (intervalId != null) {
        clearInterval(intervalId)
        intervalId = undefined
      }
    }

    const poll = async () => {
      if (cancelled) return
      try {
        const res = await getTaskStatus(globalGenPollTaskId)
        const st = res.status || ""
        if (cancelled) return

        if (st === "completed") {
          stop()
          setGlobalGenPollTaskId(null)
          setGenerating(false)
          try {
            const mapRes = await getGlobalMindMap()
            if (!cancelled) {
              setGlobalNodes(mapRes.nodes || [])
              setGlobalEdges(mapRes.edges || [])
            }
          } catch (e) {
            console.error("Failed to refresh global mindmap", e)
          }
          if (!cancelled) alert("全局思维导图生成完成")
          return
        }
        if (st === "failed") {
          stop()
          setGlobalGenPollTaskId(null)
          setGenerating(false)
          if (!cancelled) alert("全局思维导图生成失败")
        }
      } catch (err) {
        console.error("poll global mindmap task", err)
      }
    }

    void poll()
    intervalId = setInterval(poll, 2000)
    return () => {
      cancelled = true
      stop()
    }
  }, [globalGenPollTaskId])

  const handleGenerateGlobalMindmap = async () => {
    setGenerating(true)
    try {
      const res = await generateGlobalMindMap()
      setGlobalGenPollTaskId(res.task_id)
    } catch (err) {
      console.error("Failed to generate global mindmap", err)
      alert("生成失败")
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
    if (score >= 0.7) return "#ef4444"
    if (score >= 0.4) return "#eab308"
    return "#22c55e"
  }

  const globalGraphData = useMemo(() => {
    const idSet = new Set(globalNodes.map((n) => n.id))
    return {
      nodes: globalNodes.map((n) => ({
        id: n.id,
        name: n.name,
        description: n.description,
        importance_score: n.importance_score,
        note_id: n.note_id,
        anchorReach: estimateGlobalNodeAnchorReach(n.name),
      })),
      links: globalEdges
        .filter((e) => idSet.has(e.from_id) && idSet.has(e.to_id))
        .map((e) => ({
          source: e.from_id,
          target: e.to_id,
          label: e.label,
        })),
    }
  }, [globalNodes, globalEdges])

  const globalGraphStats = useMemo(
    () => ({
      nNodes: globalGraphData.nodes.length,
      nEdges: globalGraphData.links.length,
    }),
    [globalGraphData],
  )

  useEffect(() => {
    if (!globalFgRef.current) return
  
    // 修改 link 距离
    globalFgRef.current.d3Force("link")?.distance(100)
  
    // 修改斥力（让图更散）
    globalFgRef.current.d3Force("charge")?.strength(-200)
  }, [globalGraphData])

  useEffect(() => {
    if (globalNodes.length === 0 || !isGlobalView) return
    const t = window.setTimeout(() => {
      globalFgRef.current?.zoomToFit?.(400, 80)
    }, 350)
    return () => window.clearTimeout(t)
  }, [globalNodes.length, isGlobalView, globalEdges.length])

  const handleGlobalNodeClick = useCallback(
    (node: { note_id?: number }) => {
      if (node.note_id != null) navigate(`/mindmap/${node.note_id}`)
    },
    [navigate],
  )

  /** 沿连线方向把箭头放在「目标节点块」之前，避免被圆角卡片遮挡 */
  const globalLinkArrowRelPos = useCallback((link: any) => {
    const t = link.target
    const s = link.source
    if (!t || !s || typeof t !== "object" || typeof s !== "object") return 0.82
    const tx = t.x as number
    const ty = t.y as number
    const sx = s.x as number
    const sy = s.y as number
    if ([tx, ty, sx, sy].some((v) => v == null || Number.isNaN(v))) return 0.82
    const dist = Math.hypot(tx - sx, ty - sy)
    if (dist < 1e-6) return 0.5
    const reach = typeof t.anchorReach === "number" ? t.anchorReach : 52
    const rel = 1 - reach / dist
    if (dist < reach * 1.25) return 0.52
    return Math.max(0.25, Math.min(0.94, rel))
  }, [])

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
    <div className="h-full min-h-0 overflow-hidden p-6 box-border">
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
                        <p className="font-semibold truncate hover:underline inline-block max-w-full" onClick={(e) => {e.stopPropagation(); navigate(`/note/${note.id}`)}}>{note.title}</p>
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
                    <div className="flex flex-col gap-3">
                      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                        <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
                          <span className="inline-flex items-center gap-1.5">
                            <Hand className="size-3.5 shrink-0 opacity-80" aria-hidden />
                            拖动画布平移
                          </span>
                          <span className="inline-flex items-center gap-1.5">
                            <ZoomIn className="size-3.5 shrink-0 opacity-80" aria-hidden />
                            滚轮缩放
                          </span>
                          <span className="inline-flex items-center gap-1.5">
                            <MousePointerClick className="size-3.5 shrink-0 opacity-80" aria-hidden />
                            点击节点进入笔记导图
                          </span>
                        </div>
                        <div className="flex flex-wrap items-center gap-3 text-xs">
                          <span className="tabular-nums text-muted-foreground">
                            {globalGraphStats.nNodes} 个知识点 · {globalGraphStats.nEdges} 条关联
                          </span>
                          <div className="flex items-center gap-2 border-l border-border pl-3">
                            <span className="text-muted-foreground">重要度</span>
                            <span className="inline-flex items-center gap-1">
                              <span className="size-2.5 rounded-full bg-[#22c55e]" title="低" />
                              <span className="text-muted-foreground">低</span>
                            </span>
                            <span className="inline-flex items-center gap-1">
                              <span className="size-2.5 rounded-full bg-[#eab308]" title="中" />
                              <span className="text-muted-foreground">中</span>
                            </span>
                            <span className="inline-flex items-center gap-1">
                              <span className="size-2.5 rounded-full bg-[#ef4444]" title="高" />
                              <span className="text-muted-foreground">高</span>
                            </span>
                          </div>
                        </div>
                      </div>

                      <div className="relative overflow-hidden rounded-xl border border-border/80 bg-linear-to-br from-muted/60 via-background to-muted/40 p-px shadow-inner dark:from-muted/25 dark:via-background dark:to-muted/20">
                        <div
                          className="pointer-events-none absolute inset-0 opacity-[0.35] dark:opacity-25"
                          style={{
                            backgroundImage: `radial-gradient(circle at 1px 1px, hsl(var(--border)) 1px, transparent 0)`,
                            backgroundSize: "20px 20px",
                          }}
                        />
                        <div className="relative h-[calc(100vh-19rem)] min-h-[320px] w-full overflow-hidden rounded-[11px] bg-background/40 dark:bg-background/60 [&_canvas]:outline-none">
                          <ForceGraph2D
                            ref={globalFgRef}
                            graphData={globalGraphData}
                            backgroundColor="transparent"
                            linkColor={() => "rgba(100, 116, 139, 0.42)"}
                            linkDirectionalArrowLength={5}
                            linkDirectionalArrowRelPos={globalLinkArrowRelPos}
                            linkDirectionalArrowColor={() => "rgba(100, 116, 139, 0.55)"}
                            linkCurvature={0.14}
                            linkWidth={1.15}
                            linkLabel={(link: any) => String(link.label || "")}
                            nodeLabel={(n: any) =>
                              [
                                n.name,
                                n.description,
                                `重要度 ${(((n.importance_score ?? 0) as number) * 100).toFixed(0)}%`,
                                n.note_id != null ? `笔记 #${n.note_id}` : "",
                              ]
                                .filter(Boolean)
                                .join("\n")
                            }
                            nodeCanvasObject={(node: any, ctx, globalScale) => {
                              paintGlobalMindMapNode(node, ctx, globalScale, getNodeColor)
                            }}
                            nodePointerAreaPaint={(node: any, color, ctx, globalScale) => {
                              paintGlobalMindMapNode(node, ctx, globalScale, getNodeColor, color)
                            }}
                            onNodeClick={(node: any) => handleGlobalNodeClick(node)}
                            cooldownTicks={160}
                          />
                        </div>
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
