import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { semanticSearch, type SearchResult } from "@/services/searchService"

export default function SearchPage() {
  const [query, setQuery] = useState("")
  const [results, setResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [hasSearched, setHasSearched] = useState(false)

  const handleSearch = async () => {
    if (!query.trim()) return

    setSearching(true)
    setHasSearched(true)
    try {
      const res = await semanticSearch({
        query: query.trim(),
        top_k: 20
      })
      setResults(res.results || [])
    } catch (err) {
      console.error("Search failed", err)
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
          <CardContent className="pt-6">
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
