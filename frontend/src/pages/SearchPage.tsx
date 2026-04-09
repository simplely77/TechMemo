import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getSearchHistory, semanticSearch, type SearchHistoryItem, type SearchResult } from "@/services/searchService"
import { getNotes } from "@/services/noteService"
import { getKnowledgePoints } from "@/services/knowledgeService"
import { useNavigate } from "react-router-dom"
import { Clock } from "lucide-react"

type SearchType = "note" | "knowledge"

function startOfDay(d: Date): number {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
}

/** 与「今天」相差的天数（0=今天，1=昨天） */
function dayDiffFromToday(iso: string): number {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return 999
  const now = new Date()
  return Math.round((startOfDay(now) - startOfDay(d)) / 86400000)
}

/** 分组标题：今天 / 昨天 / 具体日期 */
function historyDayGroupLabel(iso: string): string {
  const diff = dayDiffFromToday(iso)
  if (diff === 0) return "今天"
  if (diff === 1) return "昨天"
  const d = new Date(iso)
  const now = new Date()
  if (d.getFullYear() === now.getFullYear()) {
    return `${d.getMonth() + 1}月${d.getDate()}日`
  }
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
}

/** 行右侧时间：今天只显示时刻，其余带日期感 */
function formatHistoryTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  const t = d.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", hour12: false })
  const diff = dayDiffFromToday(iso)
  if (diff === 0) return t
  if (diff === 1) return `昨天 ${t}`
  if (d.getFullYear() === new Date().getFullYear()) {
    return `${d.getMonth() + 1}/${d.getDate()} ${t}`
  }
  return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()} ${t}`
}

function groupSearchHistoryByDay(items: SearchHistoryItem[]): { label: string; items: SearchHistoryItem[] }[] {
  const sorted = [...items].sort(
    (a, b) => new Date(b.last_searched_at).getTime() - new Date(a.last_searched_at).getTime(),
  )
  const groups: { label: string; items: SearchHistoryItem[] }[] = []
  for (const item of sorted) {
    const label = historyDayGroupLabel(item.last_searched_at)
    const last = groups[groups.length - 1]
    if (last && last.label === label) last.items.push(item)
    else groups.push({ label, items: [item] })
  }
  return groups
}

export default function SearchPage() {
  const navigate = useNavigate()
  const [query, setQuery] = useState("")
  const [results, setResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [hasSearched, setHasSearched] = useState(false)
  const [searchType, setSearchType] = useState<SearchType>("note")
  const [enabled, setEnabled] = useState(false)
  const [searchHistory, setSearchHistory] = useState<SearchHistoryItem[]>([])

  useEffect(() => {
    getSearchHistory(1, 20, enabled ? "semantic" : "keyword", searchType).then((res) => {
      setSearchHistory(res.items)
    })
    setResults([])
    setQuery("")
    setHasSearched(false)
  }, [searchType, enabled])
  
  const handleSearch = async (keywordOverride?: string) => {
    const raw = keywordOverride !== undefined ? keywordOverride : query
    const trimmed = raw.trim()
    if (!trimmed) return
    if (keywordOverride !== undefined) setQuery(trimmed)

    setSearching(true)
    setHasSearched(true)
    try {
      let searchResults: SearchResult[] = []

      if (enabled) {
        // 语义搜索
        const res = await semanticSearch({
          query: trimmed,
          search_type: searchType,
          top_k: 20
        })
        searchResults = res.results || []
      } else {
        // 关键词搜索
        if (searchType === "note") {
          const res = await getNotes({
            keyword: trimmed,
            page: 1,
            page_size: 20
          })
          // 将 Note 转换为 SearchResult
          searchResults = (res.notes || []).map(note => ({
            id: note.id,
            type: "note",
            title: note.title,
            content: note.content_md,
            similarity: 1, // 关键词搜索没有相似度分数
            created_at: note.created_at,
            note_type: note.note_type,
            category: note.category
          }))
        } else {
          const res = await getKnowledgePoints({
            keyword: trimmed,
            page: 1,
            page_size: 20
          })
          // 将 KnowledgePoint 转换为 SearchResult
          searchResults = (res.knowledge_points || []).map(kp => ({
            id: kp.id,
            type: "knowledge",
            title: kp.name,
            content: kp.description,
            similarity: 1, // 关键词搜索没有相似度分数
            created_at: kp.created_at,
            source_note_id: kp.source_note_id,
            source_note_title: kp.source_note_title,
            importance_score: kp.importance_score
          }))
        }
      }

      setResults(searchResults)
      getSearchHistory(1, 20, enabled ? "semantic" : "keyword", searchType)
        .then((res) => setSearchHistory(res.items))
        .catch(() => {})
      console.log("searchHistory", searchHistory)
    } catch (err) {
      console.error("搜索失败:", err)
      setResults([])
    } finally {
      setSearching(false)
    }
  }

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">语义搜索</h1>

        {/* 搜索框 */}
        <Card className="mb-6">
          <CardContent className="pt-1">
            <div className="space-y-5">
              {/* 搜索类型选择 */}
              <div className="flex gap-4 items-center">
                <button
                  onClick={() => setSearchType("note")}
                  className={`px-4 py-2 rounded text-sm transition-colors ${searchType === "note"
                    ? "bg-primary text-primary-foreground"
                    : "bg-accent hover:bg-accent/80"
                    }`}
                >
                  笔记
                </button>
                <button
                  onClick={() => setSearchType("knowledge")}
                  className={`px-4 py-2 rounded text-sm transition-colors ${searchType === "knowledge"
                    ? "bg-primary text-primary-foreground"
                    : "bg-accent hover:bg-accent/80"
                    }`}
                >
                  知识点
                </button>
                <div className="ml-auto flex items-center gap-2">
                  <span className={!enabled ? "font-medium" : "text-muted-foreground"}>
                    关键词
                  </span>
                  <Switch checked={enabled} onCheckedChange={setEnabled} className="bg-muted data-[state=checked]:bg-muted [&>span]:bg-white data-[state=checked]:[&>span]:bg-white" />
                  <span className={enabled ? "font-medium" : "text-muted-foreground"}>
                    语义
                  </span>
                </div>
              </div>

              {/* 搜索输入框 */}
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="输入想要搜索的内容..."
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && !searching && void handleSearch()}
                  className="flex-1 px-3 py-2 border rounded"
                />
                <Button onClick={() => void handleSearch()} disabled={searching || !query.trim()}>
                  {searching ? "搜索中..." : "搜索"}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 未搜索：浏览器式历史列表；已搜索：本次结果 */}
        <div className="space-y-4">
          {hasSearched && (
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-8 px-2 text-muted-foreground hover:text-foreground"
                onClick={() => setHasSearched(false)}
              >
                ← 搜索历史
              </Button>
            </div>
          )}
          {!hasSearched ? (
            <div className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
              <div className="border-b border-border bg-muted/40 px-4 py-2.5 text-sm font-medium text-foreground">
                搜索历史
              </div>
              {searchHistory.length === 0 ? (
                <div className="px-4 py-12 text-center text-sm text-muted-foreground">
                  暂无记录。在上方输入内容并搜索后，会显示在这里。
                </div>
              ) : (
                <div className="divide-y divide-border">
                  {groupSearchHistoryByDay(searchHistory).map((group) => (
                    <div key={group.label}>
                      <div className="bg-muted/30 px-4 py-1.5 text-xs font-medium text-muted-foreground">
                        {group.label}
                      </div>
                      <ul className="divide-y divide-border">
                        {group.items.map((item) => (
                          <li key={item.id}>
                            <button
                              type="button"
                              disabled={searching}
                              className="flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-muted/70 disabled:pointer-events-none disabled:opacity-50"
                              onClick={() => void handleSearch(item.keyword)}
                            >
                              <Clock
                                className="size-4 shrink-0 text-muted-foreground opacity-70"
                                strokeWidth={1.75}
                                aria-hidden
                              />
                              <div className="min-w-0 flex-1">
                                <div className="truncate text-sm text-foreground">{item.keyword}</div>
                                <div className="truncate text-xs text-muted-foreground">
                                  {item.search_type === "semantic" ? "语义" : "关键词"} ·{" "}
                                  {item.target_type === "note" ? "笔记" : "知识点"}
                                </div>
                              </div>
                              <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
                                {formatHistoryTime(item.last_searched_at)}
                              </span>
                            </button>
                          </li>
                        ))}
                      </ul>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : results.length === 0 ? (
            <Card>
              <CardContent className="py-12">
                <p className="text-center text-muted-foreground">未找到相关结果</p>
              </CardContent>
            </Card>
          ) : (
            [...results]
              .sort((a, b) => b.similarity - a.similarity)
              .map((result) => (
                <Card
                  key={`${result.type}-${result.id}`}
                  className="cursor-pointer transition-all hover:shadow-lg"
                >
                  <CardHeader>
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <CardTitle
                          className="text-lg hover:underline hover:text-blue-500"
                          onClick={() =>
                            result.type === "note"
                              ? navigate(`/note/${result.id}`)
                              : navigate(`/knowledge/${result.id}`)
                          }
                        >
                          {result.title}
                        </CardTitle>
                        <p className="text-xs text-muted-foreground mt-1">
                          {result.type === "note" ? "笔记" : "知识点"} · 相关度:{" "}
                          {(result.similarity * 100).toFixed(0)}%
                        </p>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground line-clamp-3">{result.content}</p>
                  </CardContent>
                </Card>
              ))
          )}
        </div>
      </div>
    </div>
  )
}
