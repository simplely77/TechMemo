import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { getCategories, createCategory, deleteCategory, type Category } from "@/services/categoryService"
import { getNotes, getNote, createNote, updateNote, deleteNote, updateNoteTags, type Note, type NoteTag } from "@/services/noteService"
import { getTags, createTag, type Tag } from "@/services/tagService"

type ViewMode = 'view' | 'edit' | 'create'

export default function NotePage() {
    const [categories, setCategories] = useState<Category[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [allTags, setAllTags] = useState<Tag[]>([])
    const [selectedCategoryId, setSelectedCategoryId] = useState<number | null>(null)
    const [selectedNote, setSelectedNote] = useState<Note | null>(null)
    const [viewMode, setViewMode] = useState<ViewMode>('view')
    const [formData, setFormData] = useState({ title: '', content: '', tagIds: [] as number[] })
    const [newCategoryName, setNewCategoryName] = useState('')
    const [showCategoryForm, setShowCategoryForm] = useState(false)
    // 笔记列表中正在编辑标签的笔记 id
    const [editingTagsNoteId, setEditingTagsNoteId] = useState<number | null>(null)
    // 新建标签输入
    const [newTagName, setNewTagName] = useState('')

    useEffect(() => {
        loadCategories()
        loadAllTags()
    }, [])

    useEffect(() => {
        if (selectedCategoryId) loadNotes(selectedCategoryId)
    }, [selectedCategoryId])

    const loadCategories = async () => {
        try {
            const res = await getCategories()
            const cats = res.categories || []
            setCategories(cats)
            if (cats.length > 0) setSelectedCategoryId(cats[0].id)
        } catch (err) { console.error(err) }
    }

    const loadAllTags = async () => {
        try {
            const res = await getTags()
            setAllTags(res.tags || [])
        } catch (err) { console.error(err) }
    }

    const loadNotes = async (categoryId: number) => {
        try {
            const res = await getNotes({ category_id: categoryId })
            const sorted = (res.notes || []).sort((a, b) =>
                new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
            )
            setNotes(sorted)
            setSelectedNote(null)
            setViewMode('view')
        } catch (err) { console.error(err) }
    }

    const handleSelectNote = async (note: Note) => {
        try {
            const full = await getNote(note.id)
            setSelectedNote(full)
            setViewMode('view')
            setEditingTagsNoteId(null)
        } catch (err) { console.error(err) }
    }

    const handleCreateNote = async () => {
        if (!formData.title.trim() || !formData.content.trim() || !selectedCategoryId) return
        try {
            const note = await createNote({
                title: formData.title,
                content_md: formData.content,
                category_id: selectedCategoryId,
                tag_ids: formData.tagIds
            })
            setNotes(prev => [note, ...prev])
            setSelectedNote(note)
            setViewMode('view')
            setFormData({ title: '', content: '', tagIds: [] })
        } catch (err) { console.error(err) }
    }

    const handleUpdateNote = async () => {
        if (!selectedNote) return
        try {
            const updated = await updateNote(selectedNote.id, {
                title: formData.title,
                content_md: formData.content
            })
            await updateNoteTags(selectedNote.id, formData.tagIds)
            // 合并 tags（updateNote 返回的对象可能没有 tags，手动补上）
            const updatedWithTags = {
                ...updated,
                tags: allTags.filter(t => formData.tagIds.includes(t.id)) as NoteTag[]
            }
            setNotes(prev => prev.map(n => n.id === updated.id ? updatedWithTags : n))
            setSelectedNote(updatedWithTags)
            setViewMode('view')
        } catch (err) { console.error(err) }
    }

    const handleDeleteNote = async () => {
        if (!selectedNote) return
        try {
            await deleteNote(selectedNote.id)
            setNotes(prev => prev.filter(n => n.id !== selectedNote.id))
            setSelectedNote(null)
        } catch (err) { console.error(err) }
    }

    const handleCreateCategory = async () => {
        if (!newCategoryName.trim()) return
        try {
            const cat = await createCategory({ name: newCategoryName })
            setCategories(prev => [...prev, cat])
            setNewCategoryName('')
            setShowCategoryForm(false)
            setSelectedCategoryId(cat.id)
            setNotes([])
            setSelectedNote(null)
        } catch (err) { console.error(err) }
    }

    const handleDeleteCategory = async (id: number) => {
        try {
            await deleteCategory(id)
            setCategories(prev => prev.filter(c => c.id !== id))
            if (selectedCategoryId === id) {
                setSelectedCategoryId(categories.find(c => c.id !== id)?.id ?? null)
            }
        } catch (err) { console.error(err) }
    }

    // 切换笔记的某个标签（直接调接口更新）
    const handleToggleNoteTag = async (note: Note, tagId: number) => {
        const currentIds = note.tags.map(t => t.id)
        const newIds = currentIds.includes(tagId)
            ? currentIds.filter(id => id !== tagId)
            : [...currentIds, tagId]
        try {
            await updateNoteTags(note.id, newIds)
            const updatedTags = allTags.filter(t => newIds.includes(t.id)) as NoteTag[]
            const updatedNote = { ...note, tags: updatedTags }
            setNotes(prev => prev.map(n => n.id === note.id ? updatedNote : n))
            if (selectedNote?.id === note.id) setSelectedNote(updatedNote)
        } catch (err) { console.error(err) }
    }

    const handleCreateAndAddTag = async (note: Note) => {
        if (!newTagName.trim()) return
        try {
            const tag = await createTag({ name: newTagName })
            setAllTags(prev => [...prev, tag])
            setNewTagName('')
            await handleToggleNoteTag(note, tag.id)
        } catch (err) { console.error(err) }
    }

    const startEdit = () => {
        if (!selectedNote) return
        setFormData({
            title: selectedNote.title,
            content: selectedNote.content_md,
            tagIds: selectedNote.tags.map(t => t.id)
        })
        setViewMode('edit')
    }

    const startCreate = () => {
        setFormData({ title: '', content: '', tagIds: [] })
        setViewMode('create')
    }

    return (
        <div className="flex h-screen">
            {/* 左侧边栏 */}
            <div className="w-64 border-r flex flex-col">
                {/* 分类区 */}
                <div className="p-3 border-b">
                    <div className="flex justify-between items-center mb-2">
                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">分类</span>
                        <button onClick={() => setShowCategoryForm(v => !v)} className="text-sm text-primary leading-none">+</button>
                    </div>
                    {showCategoryForm && (
                        <div className="flex gap-1 mb-2">
                            <input
                                className="flex-1 text-xs border rounded px-1.5 py-1"
                                placeholder="分类名"
                                value={newCategoryName}
                                onChange={e => setNewCategoryName(e.target.value)}
                                onKeyDown={e => {
                                    if (e.key === 'Enter') handleCreateCategory()
                                    if (e.key === 'Escape') { setShowCategoryForm(false); setNewCategoryName('') }
                                }}
                                autoFocus
                            />
                            <button onClick={handleCreateCategory} className="text-xs text-primary">✓</button>
                            <button onClick={() => { setShowCategoryForm(false); setNewCategoryName('') }} className="text-xs text-muted-foreground">✕</button>
                        </div>
                    )}
                    <div className="space-y-0.5">
                        {categories.map(cat => (
                            <div
                                key={cat.id}
                                className={`flex items-center justify-between px-2 py-1 rounded text-sm cursor-pointer group ${selectedCategoryId === cat.id ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'}`}
                                onClick={() => setSelectedCategoryId(cat.id)}
                            >
                                <span className="truncate">{cat.name}</span>
                                <button
                                    className="opacity-0 group-hover:opacity-100 text-xs shrink-0 ml-1"
                                    onClick={e => { e.stopPropagation(); handleDeleteCategory(cat.id) }}
                                >✕</button>
                            </div>
                        ))}
                    </div>
                </div>

                {/* 笔记列表 */}
                <div className="flex-1 overflow-y-auto p-3">
                    <div className="flex justify-between items-center mb-2">
                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">笔记</span>
                        <button onClick={startCreate} className="text-sm text-primary leading-none">+</button>
                    </div>
                    <div className="space-y-1">
                        {notes.map(note => (
                            <div key={note.id} className="relative">
                                <div
                                    className={`px-2 py-1.5 rounded cursor-pointer group ${selectedNote?.id === note.id ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'}`}
                                    onClick={() => handleSelectNote(note)}
                                >
                                    <div className="flex items-center justify-between">
                                        <span className="text-xs truncate flex-1">{note.title}</span>
                                        <button
                                            className="opacity-0 group-hover:opacity-100 text-xs shrink-0 ml-1 px-1"
                                            onClick={e => {
                                                e.stopPropagation()
                                                setEditingTagsNoteId(editingTagsNoteId === note.id ? null : note.id)
                                            }}
                                            title="编辑标签"
                                        >
                                            🏷
                                        </button>
                                    </div>
                                    {/* 标签展示（最多2个） */}
                                    {note.tags.length > 0 && (
                                        <div className="flex gap-1 mt-1 flex-wrap">
                                            {note.tags.slice(0, 2).map(tag => (
                                                <span
                                                    key={tag.id}
                                                    className={`text-[10px] px-1.5 py-0.5 rounded-full ${selectedNote?.id === note.id ? 'bg-primary-foreground/20 text-primary-foreground' : 'bg-accent text-muted-foreground'}`}
                                                >
                                                    {tag.name}
                                                </span>
                                            ))}
                                            {note.tags.length > 2 && (
                                                <span className="text-[10px] text-muted-foreground">+{note.tags.length - 2}</span>
                                            )}
                                        </div>
                                    )}
                                </div>

                                {/* 标签编辑弹出层 */}
                                {editingTagsNoteId === note.id && (
                                    <div className="absolute left-0 right-0 z-20 mt-1 bg-background border rounded shadow-lg p-2">
                                        <div className="flex flex-wrap gap-1 mb-2">
                                            {allTags.map(tag => (
                                                <button
                                                    key={tag.id}
                                                    onClick={() => handleToggleNoteTag(note, tag.id)}
                                                    className={`text-[10px] px-1.5 py-0.5 rounded-full border transition-colors ${note.tags.some(t => t.id === tag.id) ? 'bg-primary text-primary-foreground border-primary' : 'hover:bg-accent'}`}
                                                >
                                                    {tag.name}
                                                </button>
                                            ))}
                                        </div>
                                        <div className="flex gap-1">
                                            <input
                                                className="flex-1 text-[10px] border rounded px-1.5 py-0.5"
                                                placeholder="新建标签..."
                                                value={newTagName}
                                                onChange={e => setNewTagName(e.target.value)}
                                                onKeyDown={e => {
                                                    if (e.key === 'Enter') handleCreateAndAddTag(note)
                                                    if (e.key === 'Escape') setEditingTagsNoteId(null)
                                                }}
                                            />
                                            <button onClick={() => handleCreateAndAddTag(note)} className="text-[10px] text-primary">+</button>
                                            <button onClick={() => setEditingTagsNoteId(null)} className="text-[10px] text-muted-foreground">✕</button>
                                        </div>
                                    </div>
                                )}
                            </div>
                        ))}
                        {notes.length === 0 && (
                            <p className="text-xs text-muted-foreground py-2">暂无笔记</p>
                        )}
                    </div>
                </div>
            </div>

            {/* 右侧内容区 */}
            <div className="flex-1 p-6 overflow-y-auto">
                {viewMode === 'create' || viewMode === 'edit' ? (
                    <Card className="h-full flex flex-col">
                        <CardContent className="flex-1 flex flex-col pt-6 gap-3">
                            <input
                                className="border rounded px-3 py-2 text-sm"
                                placeholder="笔记标题"
                                value={formData.title}
                                onChange={e => setFormData(p => ({ ...p, title: e.target.value }))}
                            />
                            <textarea
                                className="flex-1 border rounded px-3 py-2 text-sm resize-none min-h-75"
                                placeholder="笔记内容（Markdown）"
                                value={formData.content}
                                onChange={e => setFormData(p => ({ ...p, content: e.target.value }))}
                            />
                            {allTags.length > 0 && (
                                <div>
                                    <p className="text-xs text-muted-foreground mb-1">选择标签</p>
                                    <div className="flex flex-wrap gap-1">
                                        {allTags.map(tag => (
                                            <button
                                                key={tag.id}
                                                onClick={() => setFormData(p => ({
                                                    ...p,
                                                    tagIds: p.tagIds.includes(tag.id)
                                                        ? p.tagIds.filter(id => id !== tag.id)
                                                        : [...p.tagIds, tag.id]
                                                }))}
                                                className={`text-xs rounded-full px-2 py-0.5 border transition-colors ${formData.tagIds.includes(tag.id) ? 'bg-primary text-primary-foreground border-primary' : 'hover:bg-accent'}`}
                                            >
                                                {tag.name}
                                            </button>
                                        ))}
                                    </div>
                                </div>
                            )}
                            <div className="flex gap-2">
                                <Button size="sm" onClick={viewMode === 'create' ? handleCreateNote : handleUpdateNote}>
                                    {viewMode === 'create' ? '创建' : '保存'}
                                </Button>
                                <Button size="sm" variant="outline" onClick={() => setViewMode('view')}>取消</Button>
                            </div>
                        </CardContent>
                    </Card>
                ) : selectedNote ? (
                    <Card className="h-full flex flex-col">
                        <CardContent className="flex-1 flex flex-col pt-6 overflow-hidden">
                            <h2 className="text-lg font-semibold mb-1">{selectedNote.title}</h2>
                            <div className="flex items-center gap-2 mb-4">
                                <p className="text-xs text-muted-foreground">
                                    更新于 {new Date(selectedNote.updated_at).toLocaleString()}
                                </p>
                                {selectedNote.tags.length > 0 && (
                                    <div className="flex gap-1 flex-wrap">
                                        {selectedNote.tags.map(tag => (
                                            <span key={tag.id} className="text-xs bg-accent px-2 py-0.5 rounded-full text-muted-foreground">
                                                {tag.name}
                                            </span>
                                        ))}
                                    </div>
                                )}
                            </div>
                            <div className="flex-1 overflow-y-auto mb-4">
                                <pre className="whitespace-pre-wrap text-sm font-sans">{selectedNote.content_md}</pre>
                            </div>
                            <div className="flex gap-2">
                                <Button size="sm" onClick={startEdit}>编辑</Button>
                                <Button size="sm" variant="destructive" onClick={handleDeleteNote}>删除</Button>
                            </div>
                        </CardContent>
                    </Card>
                ) : (
                    <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
                        选择一篇笔记，或点击 + 新建
                    </div>
                )}
            </div>
        </div>
    )
}
