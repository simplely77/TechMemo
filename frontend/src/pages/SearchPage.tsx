import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { semanticSearch, type SearchResult } from "@/services/searchService"
import { getNotes } from "@/services/noteService"
import { getKnowledgePoints } from "@/services/knowledgeService"

type SearchType = "note" | "knowledge"

export default function SearchPage() {
  const [query, setQuery] = useState("")
  const [results, setResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [hasSearched, setHasSearched] = useState(false)
  const [searchType, setSearchType] = useState<SearchType>("note")
  const [enabled, setEnabled] = useState(false)

  const handleSearch = async () => {
    if (!query.trim()) return

    setSearching(true)
    setHasSearched(true)
    try {
      let searchResults: SearchResult[] = []

      if (enabled) {
        // 语义搜索
        const res = await semanticSearch({
          query: query.trim(),
          search_type: searchType,
          top_k: 20
        })
        searchResults = res.results || []
      } else {
        // 关键词搜索
        if (searchType === "note") {
          const res = await getNotes({
            keyword: query.trim(),
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
            keyword: query.trim(),
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
                  onKeyPress={(e) => e.key === "Enter" && !searching && handleSearch()}
                  className="flex-1 px-3 py-2 border rounded"
                />
                <Button onClick={handleSearch} disabled={searching || !query.trim()}>
                  {searching ? "搜索中..." : "搜索"}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 搜索结果 */}
        <div className="space-y-4">
          {!hasSearched ? (
            <Card>
              <CardContent className="p-6">
                <p className="text-muted-foreground text-center text-3xl">输入关键词开始搜索</p>
              </CardContent>
            </Card>
          ) : results.length === 0 ? (
            <Card>
              <CardContent className="p-6">
                <p className="text-muted-foreground text-center text-3xl">未找到相关结果</p>
              </CardContent>
            </Card>
          ) : (
            results.map(result => (
              <Card key={result.id} className="cursor-pointer hover:shadow-lg transition-all">
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <CardTitle className="text-lg">{result.title}</CardTitle>
                      <p className="text-xs text-muted-foreground mt-1">
                        {result.type === "note" ? "笔记" : "知识点"} · 相关度: {(result.similarity * 100).toFixed(0)}%
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
