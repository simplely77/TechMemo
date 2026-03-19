import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { getCategories, type Category } from "@/services/categoryService"
import { getNotes, type Note } from "@/services/noteService"

export default function NotePage() {
    const [sidebarOpen, setSidebarOpen] = useState(true)
    const [categories, setCategories] = useState<Category[]>([])
    const [selectedCategory, setSelectedCategory] = useState<Category | null>(null)
    const [notes, setNotes] = useState<Note[]>([])

    useEffect(() => {
        getCategories()
            .then(res => {
                const cats = res.categories
                setCategories(cats)
                if (cats.length > 0) setSelectedCategory(cats[0])
            })
            .catch(err => console.error("Failed to fetch categories", err))
    }, [])

    useEffect(() => {
        if (!selectedCategory) return
        getNotes({ category_id: selectedCategory.id })
            .then(res => setNotes(res.notes))
            .catch(err => console.error("Failed to fetch categories", err))
    }, [selectedCategory])

    return (
        <div className="flex h-screen">
            {/* 左侧 */}
            <div className={`border-r p-4 flex flex-col ${sidebarOpen ? "w-64" : "w-16"}`}>
                <div className="flex justify-between items-center mb-4">
                    {sidebarOpen && <h2 className="text-xl font-bold">笔记管理</h2>}
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSidebarOpen(!sidebarOpen)}
                    >
                        {sidebarOpen ? "◀" : "▶"}
                    </Button>
                </div>
                {sidebarOpen && (
                    <ul className="flex flex-col gap-1">
                        {categories.map((cat) => (
                            <li key={cat.id} className="p-2 transition-all hover:shadow-lg hover:-translate-y-1 rounded" onClick={() => setSelectedCategory(cat)}>{cat.name}</li>
                        ))}
                    </ul>
                )}
            </div>
            {/* 右侧 */}
            <div className="flex-1 p-6">
                <h2 className="text-lg font-bold mb-4">
                    {selectedCategory ? selectedCategory.name : "请选择分类"}
                </h2>

                {notes.length === 0 ? (
                    <p>该分类下暂无笔记</p>
                ) : (
                    <ul className="space-y-2">
                        {notes.map(note => (
                            <li key={note.id} className="p-2 border rounded hover:shadow">
                                <h3 className="font-semibold">{note.title}</h3>
                            </li>
                        ))}
                    </ul>
                )}
            </div>
        </div>
    )
}