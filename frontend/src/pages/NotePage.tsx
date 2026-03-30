import { useEffect, useState } from "react"
import { Card, CardContent } from "@/components/ui/card"
import { getCategories, createCategory, deleteCategory, type Category } from "@/services/categoryService"
import { getNotes, getNote, createNote, updateNote, deleteNote, updateNoteTags, type Note, type NoteTag } from "@/services/noteService"
import { getTags, createTag, deleteTag, type Tag } from "@/services/tagService"
import { processNoteAI, getNoteAIStatus } from "@/services/aiService"
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import remarkBreaks from 'remark-breaks'
import rehypeKatex from 'rehype-katex'
import rehypeHighlight from 'rehype-highlight'
import rehypeSlug from 'rehype-slug'
import rehypeAutolinkHeadings from 'rehype-autolink-headings'
import rehypeRaw from 'rehype-raw'
import 'katex/dist/katex.min.css'
import 'highlight.js/styles/github-dark.css'

type ViewMode = 'view' | 'edit'

export default function NotePage() {
    const [categories, setCategories] = useState<Category[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [allTags, setAllTags] = useState<Tag[]>([])
    const [selectedCategoryId, setSelectedCategoryId] = useState<number | null>(null)
    const [selectedNote, setSelectedNote] = useState<Note | null>(null)
    const [viewMode, setViewMode] = useState<ViewMode>('view')
    const [editingContent, setEditingContent] = useState('')
    const [newCategoryName, setNewCategoryName] = useState('')
    const [showCategoryForm, setShowCategoryForm] = useState(false)
    const [newNoteName, setNewNoteName] = useState('')
    const [showNoteForm, setShowNoteForm] = useState(false)
    // 笔记列表中正在编辑标签的笔记 id
    const [editingTagsNoteId, setEditingTagsNoteId] = useState<number | null>(null)
    // 笔记列表中正在重命名的笔记 id
    const [renamingNoteId, setRenamingNoteId] = useState<number | null>(null)
    const [renamingTitle, setRenamingTitle] = useState('')
    // 新建标签输入
    const [newTagName, setNewTagName] = useState('')
    // 标签筛选
    const [selectedTagIds, setSelectedTagIds] = useState<number[]>([])
    const [showTagForm, setShowTagForm] = useState(false)
    const [newTagNameInSidebar, setNewTagNameInSidebar] = useState('')
    // AI 处理状态
    const [aiProcessing, setAiProcessing] = useState(false)
    const [aiStatus, setAiStatus] = useState<string>('')

    useEffect(() => {
        loadCategories()
        loadAllTags()
    }, [])

    useEffect(() => {
        if (selectedCategoryId) loadNotes(selectedCategoryId)
    }, [selectedCategoryId])

    // 轮询 AI 处理状态
    useEffect(() => {
        if (!selectedNote || !aiProcessing) return

        const pollStatus = async () => {
            try {
                const status = await getNoteAIStatus(selectedNote.id)
                setAiStatus(status.status || '')

                // 如果处理完成，停止轮询
                if (status.status === 'completed' || status.status === 'failed') {
                    setAiProcessing(false)
                    if (status.status === 'completed') {
                        alert('AI 处理完成')
                    } else {
                        alert('AI 处理失败')
                    }
                }
            } catch (err) {
                console.error(err)
            }
        }

        const interval = setInterval(pollStatus, 2000) // 每2秒轮询一次
        return () => clearInterval(interval)
    }, [selectedNote, aiProcessing])

    // 键盘快捷键：保存
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            // Cmd/Ctrl + S: 保存
            if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                e.preventDefault()
                if (viewMode === 'edit') {
                    handleUpdateNote()
                }
            }
            // Escape: 退出编辑模式并保存
            if (e.key === 'Escape' && viewMode === 'edit') {
                e.preventDefault()
                handleClickOutside()
            }
        }

        window.addEventListener('keydown', handleKeyDown)
        return () => window.removeEventListener('keydown', handleKeyDown)
    }, [viewMode, editingContent, selectedNote])

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
        // 如果正在编辑且有未保存的内容，先保存
        if (viewMode === 'edit' && selectedNote && editingContent !== selectedNote.content_md) {
            await handleUpdateNote()
        }

        try {
            const full = await getNote(note.id)
            setSelectedNote(full)
            setViewMode('view')
            setEditingTagsNoteId(null)
            setEditingContent('')
        } catch (err) { console.error(err) }
    }

    const handleCreateNote = async () => {
        if (!newNoteName.trim() || !selectedCategoryId) return
        try {
            const note = await createNote({
                title: newNoteName,
                content_md: '',
                category_id: selectedCategoryId,
                tag_ids: []
            })
            setNotes(prev => [note, ...prev])
            setSelectedNote(note)
            setNewNoteName('')
            setShowNoteForm(false)
            // 自动进入编辑模式
            setEditingContent('')
            setViewMode('edit')
        } catch (err) { console.error(err) }
    }

    const handleUpdateNote = async () => {
        if (!selectedNote) return
        try {
            const updated = await updateNote(selectedNote.id, {
                title: selectedNote.title,
                content_md: editingContent
            })
            const updatedWithTags = {
                ...updated,
                tags: selectedNote.tags
            }
            setNotes(prev => prev.map(n => n.id === updated.id ? updatedWithTags : n))
            setSelectedNote(updatedWithTags)
            setViewMode('view')
        } catch (err) { console.error(err) }
    }

    const handleRenameNote = async (note: Note, newTitle: string) => {
        if (!newTitle.trim()) return
        try {
            const updated = await updateNote(note.id, {
                title: newTitle,
                content_md: note.content_md
            })
            const updatedWithTags = {
                ...updated,
                tags: note.tags
            }
            setNotes(prev => prev.map(n => n.id === note.id ? updatedWithTags : n))
            if (selectedNote?.id === note.id) {
                setSelectedNote(updatedWithTags)
            }
            setRenamingNoteId(null)
            setRenamingTitle('')
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

    // 在侧边栏创建新标签
    const handleCreateTagInSidebar = async () => {
        if (!newTagNameInSidebar.trim()) return
        try {
            const tag = await createTag({ name: newTagNameInSidebar })
            setAllTags(prev => [...prev, tag])
            setNewTagNameInSidebar('')
            setShowTagForm(false)
        } catch (err) { console.error(err) }
    }

    // 删除标签
    const handleDeleteTag = async (id: number) => {
        try {
            await deleteTag(id)
            setAllTags(prev => prev.filter(t => t.id !== id))
            setSelectedTagIds(prev => prev.filter(tid => tid !== id))
        } catch (err) { console.error(err) }
    }

    // 切换标签筛选
    const handleToggleTagFilter = (tagId: number) => {
        setSelectedTagIds(prev =>
            prev.includes(tagId)
                ? prev.filter(id => id !== tagId)
                : [...prev, tagId]
        )
    }

    // AI 处理笔记
    const handleProcessAI = async () => {
        if (!selectedNote) return
        setAiProcessing(true)
        setAiStatus('processing')
        try {
            await processNoteAI(selectedNote.id)
        } catch (err) {
            console.error(err)
            alert('AI 处理失败')
            setAiProcessing(false)
            setAiStatus('')
        }
    }

    // 根据选中的标签筛选笔记
    const filteredNotes = selectedTagIds.length === 0
        ? notes
        : notes.filter(note =>
            note.tags.some(tag => selectedTagIds.includes(tag.id))
        )

    const startEdit = () => {
        if (!selectedNote) return
        setEditingContent(selectedNote.content_md)
        setViewMode('edit')
    }

    // 点击底部信息栏时，如果在编辑模式则保存退出
    const handleClickOutside = async () => {
        if (viewMode !== 'edit') return
        if (selectedNote && editingContent !== selectedNote.content_md) {
            await handleUpdateNote()
        } else {
            setViewMode('view')
        }
    }

    return (
        <div className="flex h-screen overflow-hidden">
            {/* 左侧边栏 */}
            <div className="w-64 min-w-[256px] border-r flex flex-col bg-background">
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

                {/* 标签区 */}
                <div className="p-3 border-b">
                    <div className="flex justify-between items-center mb-2">
                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">标签</span>
                        <button onClick={() => setShowTagForm(v => !v)} className="text-sm text-primary leading-none">+</button>
                    </div>
                    {showTagForm && (
                        <div className="flex gap-1 mb-2">
                            <input
                                className="flex-1 text-xs border rounded px-1.5 py-1"
                                placeholder="标签名"
                                value={newTagNameInSidebar}
                                onChange={e => setNewTagNameInSidebar(e.target.value)}
                                onKeyDown={e => {
                                    if (e.key === 'Enter') handleCreateTagInSidebar()
                                    if (e.key === 'Escape') { setShowTagForm(false); setNewTagNameInSidebar('') }
                                }}
                                autoFocus
                            />
                            <button onClick={handleCreateTagInSidebar} className="text-xs text-primary">✓</button>
                            <button onClick={() => { setShowTagForm(false); setNewTagNameInSidebar('') }} className="text-xs text-muted-foreground">✕</button>
                        </div>
                    )}
                    <div className="flex flex-wrap gap-1">
                        {allTags.map(tag => (
                            <div
                                key={tag.id}
                                className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs cursor-pointer group transition-colors ${selectedTagIds.includes(tag.id) ? 'bg-primary text-primary-foreground' : 'bg-accent hover:bg-accent/60'}`}
                                onClick={() => handleToggleTagFilter(tag.id)}
                            >
                                <span className="truncate">{tag.name}</span>
                                <button
                                    className="opacity-0 group-hover:opacity-100 text-xs shrink-0"
                                    onClick={e => { e.stopPropagation(); handleDeleteTag(tag.id) }}
                                >✕</button>
                            </div>
                        ))}
                        {allTags.length === 0 && (
                            <p className="text-xs text-muted-foreground py-1">暂无标签</p>
                        )}
                    </div>
                    {selectedTagIds.length > 0 && (
                        <div className="mt-2 flex items-center justify-between">
                            <span className="text-[10px] text-muted-foreground">已选 {selectedTagIds.length} 个标签</span>
                            <button
                                onClick={() => setSelectedTagIds([])}
                                className="text-[10px] text-primary"
                            >清除筛选</button>
                        </div>
                    )}
                </div>

                {/* 笔记列表 */}
                <div className="flex-1 overflow-y-auto p-3">
                    <div className="flex justify-between items-center mb-2">
                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">笔记</span>
                        <button onClick={() => setShowNoteForm(v => !v)} className="text-sm text-primary leading-none">+</button>
                    </div>
                    {showNoteForm && (
                        <div className="flex gap-1 mb-2">
                            <input
                                className="flex-1 text-xs border rounded px-1.5 py-1"
                                placeholder="笔记标题"
                                value={newNoteName}
                                onChange={e => setNewNoteName(e.target.value)}
                                onKeyDown={e => {
                                    if (e.key === 'Enter') handleCreateNote()
                                    if (e.key === 'Escape') { setShowNoteForm(false); setNewNoteName('') }
                                }}
                                autoFocus
                            />
                            <button onClick={handleCreateNote} className="text-xs text-primary">✓</button>
                            <button onClick={() => { setShowNoteForm(false); setNewNoteName('') }} className="text-xs text-muted-foreground">✕</button>
                        </div>
                    )}
                    <div className="space-y-1">
                        {filteredNotes.map(note => (
                            <div key={note.id} className="relative">
                                {renamingNoteId === note.id ? (
                                    <div className="flex gap-1 px-2 py-1.5">
                                        <input
                                            className="flex-1 text-xs border rounded px-1.5 py-0.5"
                                            value={renamingTitle}
                                            onChange={e => setRenamingTitle(e.target.value)}
                                            onKeyDown={e => {
                                                if (e.key === 'Enter') handleRenameNote(note, renamingTitle)
                                                if (e.key === 'Escape') { setRenamingNoteId(null); setRenamingTitle('') }
                                            }}
                                            autoFocus
                                        />
                                        <button onClick={() => handleRenameNote(note, renamingTitle)} className="text-xs text-primary">✓</button>
                                        <button onClick={() => { setRenamingNoteId(null); setRenamingTitle('') }} className="text-xs text-muted-foreground">✕</button>
                                    </div>
                                ) : (
                                    <div
                                        className={`px-2 py-1.5 rounded cursor-pointer group ${selectedNote?.id === note.id ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'}`}
                                        onClick={() => handleSelectNote(note)}
                                    >
                                        <div className="flex items-center justify-between">
                                            <span className="text-xs truncate flex-1">{note.title}</span>
                                            <div className="flex items-center gap-1">
                                                <button
                                                    className="opacity-0 group-hover:opacity-100 text-xs shrink-0 px-1"
                                                    onClick={e => {
                                                        e.stopPropagation()
                                                        setRenamingNoteId(note.id)
                                                        setRenamingTitle(note.title)
                                                    }}
                                                    title="重命名"
                                                >
                                                    ✏️
                                                </button>
                                                <button
                                                    className="opacity-0 group-hover:opacity-100 text-xs shrink-0 px-1"
                                                    onClick={e => {
                                                        e.stopPropagation()
                                                        setEditingTagsNoteId(editingTagsNoteId === note.id ? null : note.id)
                                                    }}
                                                    title="编辑标签"
                                                >
                                                    🏷
                                                </button>
                                                <button
                                                    className="opacity-0 group-hover:opacity-100 text-xs shrink-0 px-1"
                                                    onClick={async (e) => {
                                                        e.stopPropagation()
                                                        if (confirm('确定删除这篇笔记吗？')) {
                                                            try {
                                                                await deleteNote(note.id)
                                                                setNotes(prev => prev.filter(n => n.id !== note.id))
                                                                if (selectedNote?.id === note.id) {
                                                                    setSelectedNote(null)
                                                                }
                                                            } catch (err) {
                                                                console.error(err)
                                                            }
                                                        }
                                                    }}
                                                    title="删除笔记"
                                                >
                                                    🗑
                                                </button>
                                            </div>
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
                                )}
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
                        {filteredNotes.length === 0 && notes.length > 0 && (
                            <p className="text-xs text-muted-foreground py-2">没有符合筛选条件的笔记</p>
                        )}
                        {notes.length === 0 && (
                            <p className="text-xs text-muted-foreground py-2">暂无笔记</p>
                        )}
                    </div>
                </div>
            </div>

            {/* 右侧内容区 */}
            <div className="flex-1 flex flex-col overflow-hidden relative p-6">
                {selectedNote ? (
                    <>
                        {/* 右上角按钮组 - 固定在右侧内容区 */}
                        <div className="absolute top-6 right-6 z-50 flex items-center gap-2">
                            {/* 编辑/预览切换按钮 */}
                            <button
                                onClick={() => {
                                    if (viewMode === 'edit') {
                                        // 从编辑切换到预览，如果有修改则保存
                                        if (selectedNote && editingContent !== selectedNote.content_md) {
                                            handleUpdateNote()
                                        } else {
                                            setViewMode('view')
                                        }
                                    } else {
                                        // 从预览切换到编辑
                                        startEdit()
                                    }
                                }}
                                className="p-2 rounded-full bg-background/90 backdrop-blur-sm hover:bg-accent transition-colors shadow-lg border border-border"
                                title={viewMode === 'edit' ? '预览' : '编辑'}
                            >
                                <span className="text-xl">{viewMode === 'edit' ? '👁️' : '✏️'}</span>
                            </button>

                            {/* AI 分析按钮 */}
                            <button
                                onClick={handleProcessAI}
                                disabled={aiProcessing}
                                className="p-2 rounded-full bg-background/90 backdrop-blur-sm hover:bg-accent transition-colors disabled:opacity-50 shadow-lg border border-border"
                                title={aiProcessing ? `AI 处理中: ${aiStatus}` : 'AI 分析'}
                            >
                                <span className="text-xl">{aiProcessing ? '⏳' : '💡'}</span>
                            </button>
                        </div>

                        <Card className="h-full flex flex-col">
                            <CardContent className="flex-1 flex flex-col pt-6 overflow-hidden">
                                {/* 内容区域 */}
                                {viewMode === 'edit' ? (
                                    <textarea
                                        className="flex-1 w-full p-4 border-0 outline-none resize-none bg-transparent"
                                        value={editingContent}
                                        onChange={e => setEditingContent(e.target.value)}
                                        autoFocus
                                    />
                                ) : (
                                    <div className="flex-1 overflow-y-auto mb-4 p-4 prose prose-sm dark:prose-invert max-w-none">
                                        {selectedNote.content_md ?
                                            <ReactMarkdown
                                                remarkPlugins={[remarkGfm, remarkMath, remarkBreaks]}
                                                rehypePlugins={[
                                                    rehypeRaw,
                                                    rehypeKatex,
                                                    rehypeHighlight,
                                                    rehypeSlug,
                                                    [rehypeAutolinkHeadings, { behavior: 'wrap' }]
                                                ]}
                                            >
                                                {editingContent || selectedNote.content_md}
                                            </ReactMarkdown> : <div className="text-sm text-muted-foreground italic">
                                                笔记为空，点击编辑按钮开始编辑
                                            </div>
                                        }
                                    </div>
                                )}

                                {/* 底部信息栏 */}
                                <div className="flex items-center gap-2" onClick={handleClickOutside}>
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
                            </CardContent>
                        </Card>
                    </>
                ) : (
                    <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
                        选择一篇笔记，或点击 + 新建
                    </div>
                )}
            </div>
        </div>
    )
}
