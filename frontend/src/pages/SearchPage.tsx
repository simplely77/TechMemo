import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { semanticSearch, type SearchResult } from "@/services/searchService"
import { getNotes, type Note } from "@/services/noteService"
import { getKnowledgePoints, type KnowledgePoint } from "@/services/knowledgeService"

type SearchMode = "keyword" | "semantic"
type SearchType = "note" | "knowledge"

export default function SearchPage() {
  const [query, setQuery] = useState("")
  const [results, setResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [hasSearched, setHasSearched] = useState(false)
  const [searchMode, setSearchMode] = useState<SearchMode>("semantic")
  const [searchType, setSearchType] = useState<SearchType>("note")

  const handleSearch = async () => {
    if (!query.trim()) return

    console.log("=== 开始搜索 ===")
    console.log("搜索模式:", searchMode)
    console.log("搜索类型:", searchType)
    console.log("搜索关键词:", query.trim())

    setSearching(true)
    setHasSearched(true)
    try {
      if (searchMode === "semantic") {
        // 语义搜索
        console.log("调用语义搜索接口...")
        const res = await semanticSearch({
          query: query.trim(),
          search_type: searchType,
          top_k: 20
        })
        console.log("语义搜索响应:", res)
        console.log("结果数量:", res.results?.length || 0)
        setResults(res.results || [])
      } else {
        // 关键词搜索
        if (searchType === "note") {
          console.log("调用笔记关键词搜索接口...")
          const res = await getNotes({ keyword: query.trim(), page_size: 20 })
          console.log("笔记搜索响应:", res)
          console.log("笔记数量:", res.notes?.length || 0)
          const noteResults: SearchResult[] = (res.notes || []).map((note: Note) => ({
            id: note.id,
            type: "note",
            title: note.title,
            content: note.content_md.substring(0, 200),
            similarity: 1
          }))
          console.log("转换后的结果:", noteResults)
          setResults(noteResults)
        } else {
          console.log("调用知识点关键词搜索接口...")
          const res = await getKnowledgePoints({ keyword: query.trim(), page_size: 20 })
          console.log("知识点搜索响应:", res)
          console.log("知识点数量:", res.knowledge_points?.length || 0)
          const kpResults: SearchResult[] = (res.knowledge_points || []).map((kp: KnowledgePoint) => ({
            id: kp.id,
            type: "knowledge",
            title: kp.name,
            content: kp.description,
            similarity: 1
          }))
          console.log("转换后的结果:", kpResults)
          setResults(kpResults)
        }
      }
    } catch (err) {
      console.error("搜索失败:", err)
      setResults([])
    } finally {
      setSearching(false)
      console.log("=== 搜索结束 ===")
    }
  }

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">搜索</h1>

        {/* 搜索框 */}
        <Card className="mb-6">
          <CardContent className="pt-6">
            <div className="space-y-3">
              {/* 搜索模式选择 */}
              <div className="flex gap-2">
                <button
                  onClick={() => setSearchMode("keyword")}
                  className={`px-4 py-2 rounded text-sm transition-colors ${
                    searchMode === "keyword"
                      ? "bg-primary text-primary-foreground"
                      : "bg-accent hover:bg-accent/80"
                  }`}
                >
                  关键词搜索
                </button>
                <button
                  onClick={() => setSearchMode("semantic")}
                  className={`px-4 py-2 rounded text-sm transition-colors ${
                    searchMode === "semantic"
                      ? "bg-primary text-primary-foreground"
                      : "bg-accent hover:bg-accent/80"
                  }`}
                >
                  语义搜索
                </button>
              </div>

              {/* 搜索类型选择 */}
              <div className="flex gap-2">
                <button
                  onClick={() => setSearchType("note")}
                  className={`px-4 py-2 rounded text-sm transition-colors ${
                    searchType === "note"
                      ? "bg-primary text-primary-foreground"
                      : "bg-accent hover:bg-accent/80"
                  }`}
                >
                  搜索笔记
                </button>
                <button
                  onClick={() => setSearchType("knowledge")}
                  className={`px-4 py-2 rounded text-sm transition-colors ${
                    searchType === "knowledge"
                      ? "bg-primary text-primary-foreground"
                      : "bg-accent hover:bg-accent/80"
                  }`}
                >
                  搜索知识点
                </button>
              </div>

              {/* 搜索输入框 */}
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="输入搜索关键词..."
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
              <CardContent className="pt-6">
                <p className="text-muted-foreground text-center">输入关键词开始搜索</p>
              </CardContent>
            </Card>
          ) : results.length === 0 ? (
            <Card>
              <CardContent className="pt-6">
                <p className="text-muted-foreground text-center">未找到相关结果</p>
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
