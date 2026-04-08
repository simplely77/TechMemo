import { useCallback, useEffect, useRef, useState, type ComponentProps } from "react"
import {
    Check,
    Eye,
    Loader2,
    MoreVertical,
    PanelLeftClose,
    PanelLeftOpen,
    Pencil,
    Plus,
    Sparkles,
    Tag as TagIcon,
    Trash2,
    X,
} from "lucide-react"
import { getCategories, createCategory, deleteCategory, type Category } from "@/services/categoryService"
import {
    getNotes,
    getNote,
    createNote,
    updateNote,
    deleteNote,
    updateNoteTags,
    getNoteVersions,
    restoreNote,
    type Note,
    type NoteTag,
    type NoteVersion,
} from "@/services/noteService"
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
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useNavigate, useParams } from 'react-router-dom'

type ViewMode = 'view' | 'edit'

/** Obsidian 式侧栏行：左侧紫条 + 浅底，避免整块 primary 填充 */
const sidebarRowBase =
    'flex items-center justify-between gap-1 border-l-2 border-transparent pl-2 pr-1.5 py-1 text-[13px] leading-tight cursor-pointer rounded-r-sm transition-colors'
const sidebarRowInactive = 'text-sidebar-foreground hover:bg-sidebar-accent/80'
const sidebarRowActive =
    'bg-sidebar-accent text-sidebar-foreground border-l-violet-500 dark:border-l-violet-400'

const obsidianInput =
    'flex-1 rounded-sm border border-sidebar-border bg-muted/40 px-2 py-1 text-[12px] text-sidebar-foreground placeholder:text-muted-foreground focus:border-violet-500/50 focus:outline-none focus:ring-1 focus:ring-violet-500/30 dark:focus:border-violet-400/50 dark:focus:ring-violet-400/25'

const NOTE_SIDEBAR_COLLAPSED_KEY = 'techmemo-note-sidebar-collapsed'

const markdownRemarkPlugins = [remarkGfm, remarkMath, remarkBreaks]
const markdownRehypePlugins: ComponentProps<typeof ReactMarkdown>['rehypePlugins'] = [
    rehypeRaw,
    rehypeKatex,
    rehypeHighlight,
    rehypeSlug,
    [rehypeAutolinkHeadings, { behavior: 'wrap' }],
]

function NoteMarkdownBody({ source, emptyHint }: { source: string; emptyHint: string }) {
    if (!source.trim()) {
        return <div className="text-sm italic text-muted-foreground">{emptyHint}</div>
    }
    return (
        <ReactMarkdown
            remarkPlugins={markdownRemarkPlugins}
            rehypePlugins={markdownRehypePlugins}
        >
            {source}
        </ReactMarkdown>
    )
}

export default function NotePage() {
    const { id } = useParams<{ id?: string }>()
    const navigate = useNavigate()
    /** 新建笔记后 navigate 到 /note/:id，路由 effect 加载完再进入编辑态 */
    const openEditAfterLoadForNoteIdRef = useRef<number | null>(null)
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
    const [versionsSubNoteId, setVersionsSubNoteId] = useState<number | null>(null)
    const [versionsList, setVersionsList] = useState<NoteVersion[]>([])
    const [versionsLoading, setVersionsLoading] = useState(false)
    /** 仅预览：正文区只读展示该版本的 Markdown，不参与编辑/保存 */
    const [versionPreview, setVersionPreview] = useState<NoteVersion | null>(null)
    const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
        try {
            return localStorage.getItem(NOTE_SIDEBAR_COLLAPSED_KEY) === '1'
        } catch {
            return false
        }
    })

    const persistSidebarCollapsed = useCallback((collapsed: boolean) => {
        setSidebarCollapsed(collapsed)
        try {
            localStorage.setItem(NOTE_SIDEBAR_COLLAPSED_KEY, collapsed ? '1' : '0')
        } catch {
            /* ignore */
        }
    }, [])

    const toggleSidebarCollapsed = useCallback(() => {
        setSidebarCollapsed(prev => {
            const next = !prev
            try {
                localStorage.setItem(NOTE_SIDEBAR_COLLAPSED_KEY, next ? '1' : '0')
            } catch {
                /* ignore */
            }
            return next
        })
    }, [])


    useEffect(() => {
        loadCategories()
        loadAllTags()
    }, [])

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

    // ⌘B / Ctrl+B：折叠或展开侧栏（与 VS Code 一致）
    useEffect(() => {
        const onKey = (e: KeyboardEvent) => {
            if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
                e.preventDefault()
                toggleSidebarCollapsed()
            }
        }
        window.addEventListener('keydown', onKey)
        return () => window.removeEventListener('keydown', onKey)
    }, [toggleSidebarCollapsed])

    useEffect(() => {
        if (selectedNote == null) setVersionPreview(null)
    }, [selectedNote?.id])

    const loadCategories = async () => {
        try {
            const res = await getCategories()
            const cats = res.categories || []
            setCategories(cats)
            // /note：默认第一个分类并拉列表；带 :id 时由下方 effect 按路由加载，避免与 selectedCategoryId 抢状态
            if (cats.length > 0 && !id) {
                const firstId = cats[0].id
                setSelectedCategoryId(firstId)
                await loadNotes(firstId)
            }
        } catch (err) { console.error(err) }
    }

    const loadAllTags = async () => {
        try {
            const res = await getTags()
            setAllTags(res.tags || [])
        } catch (err) { console.error(err) }
    }

    const loadNotes = async (categoryId: number, opts?: { selectNoteId?: number }) => {
        try {
            const res = await getNotes({ category_id: categoryId })
            const sorted = (res.notes || []).sort((a, b) =>
                new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
            )
            setNotes(sorted)
            setVersionPreview(null)
            setViewMode('view')
            if (opts?.selectNoteId != null) {
                try {
                    const full = await getNote(opts.selectNoteId)
                    setSelectedNote(full)
                } catch {
                    setSelectedNote(null)
                }
            } else {
                setSelectedNote(null)
            }
        } catch (err) { console.error(err) }
    }

    // 路由无 id：URL 为列表态，清空当前笔记选中
    useEffect(() => {
        if (id) return
        setSelectedNote(null)
        setVersionPreview(null)
        setViewMode('view')
        setEditingContent('')
    }, [id])

    // 路由有 id：按 URL 打开对应笔记（与侧栏点击、外链、前进后退一致）
    useEffect(() => {
        if (!id) return
        const noteId = +id
        if (!Number.isFinite(noteId)) return
        let cancelled = false
        setSelectedNote(null)
        setVersionPreview(null)
        ;(async () => {
            try {
                const full = await getNote(noteId)
                if (cancelled) return
                setSelectedCategoryId(full.category.id)
                await loadNotes(full.category.id, { selectNoteId: noteId })
                if (cancelled) return
                setEditingTagsNoteId(null)
                if (openEditAfterLoadForNoteIdRef.current === noteId) {
                    openEditAfterLoadForNoteIdRef.current = null
                    setViewMode('edit')
                    setEditingContent(full.content_md)
                } else {
                    setEditingContent('')
                }
            } catch (err) {
                console.error(err)
            }
        })()
        return () => {
            cancelled = true
        }
    }, [id])

    // /note 无 id：当前分类下若有笔记，默认打开列表第一篇（与侧栏顺序一致）
    // 必须确认 notes 已属于当前分类：切换分类时 loadNotes 异步，否则会误用旧列表把路由锁回上一分类
    useEffect(() => {
        if (id) return
        if (selectedCategoryId == null) return
        if (notes.length === 0) return
        const first = notes[0]
        if (first.category.id !== selectedCategoryId) return
        navigate(`/note/${first.id}`, { replace: true })
    }, [id, selectedCategoryId, notes, navigate])

    const handleSelectNote = async (noteId: number) => {
        if (viewMode === 'edit' && selectedNote && editingContent !== selectedNote.content_md) {
            await handleUpdateNote()
        }
        setVersionPreview(null)
        setEditingTagsNoteId(null)
        navigate(`/note/${noteId}`)
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
            setNewNoteName('')
            setShowNoteForm(false)
            openEditAfterLoadForNoteIdRef.current = note.id
            navigate(`/note/${note.id}`)
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

    const handleCreateCategory = async () => {
        if (!newCategoryName.trim()) return
        try {
            const cat = await createCategory({ name: newCategoryName })
            setCategories(prev => [...prev, cat])
            setNewCategoryName('')
            setShowCategoryForm(false)
            setSelectedCategoryId(cat.id)
            setNotes([])
            navigate('/note')
            setSelectedNote(null)
            setVersionPreview(null)
        } catch (err) { console.error(err) }
    }

    const handleDeleteCategory = async (id: number) => {
        try {
            await deleteCategory(id)
            const remaining = categories.filter(c => c.id !== id)
            setCategories(remaining)
            if (selectedCategoryId === id) {
                const nextId = remaining[0]?.id ?? null
                setSelectedCategoryId(nextId)
                if (nextId != null) {
                    navigate('/note')
                    await loadNotes(nextId)
                } else {
                    setNotes([])
                    navigate('/note')
                    setSelectedNote(null)
                    setVersionPreview(null)
                }
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

    const loadVersionsForSubmenu = async (noteId: number) => {
        setVersionsSubNoteId(noteId)
        setVersionsLoading(true)
        setVersionsList([])
        try {
            const res = await getNoteVersions(noteId)
            setVersionsList(res.versions ?? [])
        } catch {
            setVersionsList([])
        } finally {
            setVersionsLoading(false)
        }
    }

    const handleMoveNoteToCategory = async (note: Note, categoryId: number) => {
        if (note.category.id === categoryId) return
        try {
            await updateNote(note.id, {
                title: note.title,
                content_md: note.content_md,
                category_id: categoryId,
            })
            setNotes(prev => prev.filter(n => n.id !== note.id))
            if (selectedNote?.id === note.id) {
                navigate('/note')
                setSelectedNote(null)
                setVersionPreview(null)
                setViewMode('view')
                setEditingContent('')
            }
        } catch (err) {
            console.error(err)
        }
    }

    const handleRestoreNoteVersion = async (note: Note, versionId: number) => {
        if (
            selectedNote?.id === note.id &&
            viewMode === 'edit' &&
            editingContent !== selectedNote.content_md
        ) {
            if (!confirm('当前有未保存修改，恢复历史版本将丢弃这些修改，是否继续？')) return
        }
        try {
            const restored = await restoreNote(note.id, versionId)
            setNotes(prev => prev.map(n => (n.id === note.id ? restored : n)))
            if (selectedNote?.id === note.id) {
                setSelectedNote(restored)
                setEditingContent(restored.content_md)
                setViewMode('view')
                setVersionPreview(null)
            }
        } catch (err) {
            console.error(err)
        }
    }

    const handlePreviewNoteVersion = async (note: Note, v: NoteVersion) => {
        if (
            selectedNote?.id === note.id &&
            viewMode === 'edit' &&
            editingContent !== selectedNote.content_md
        ) {
            if (!confirm('当前有未保存修改，预览将离开编辑区且不会自动保存，是否继续？')) return
        }
        try {
            const full = await getNote(note.id)
            setSelectedNote(full)
            setVersionPreview(v)
            setViewMode('view')
            setEditingTagsNoteId(null)
            setEditingContent('')
        } catch (err) {
            console.error(err)
        }
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
        setVersionPreview(null)
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

    /** 正文相对服务器是否有未保存修改（VS Code「脏」状态） */
    const isNoteDirty =
        viewMode === 'edit' &&
        selectedNote !== null &&
        editingContent !== selectedNote.content_md

    return (
        <div className="flex h-screen w-full overflow-hidden bg-background">
            {/* 左侧边栏 — 可折叠 */}
            <div
                className={`flex shrink-0 flex-col overflow-hidden border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width,min-width] duration-200 ease-out ${
                    sidebarCollapsed
                        ? 'w-0 min-w-0 border-r-0'
                        : 'w-[272px] min-w-[272px] border-r'
                }`}
            >
                <div className="flex h-9 shrink-0 items-center justify-end border-b border-sidebar-border px-2">
                    <button
                        type="button"
                        onClick={() => persistSidebarCollapsed(true)}
                        className="inline-flex h-7 w-7 items-center justify-center rounded-sm text-muted-foreground hover:bg-sidebar-accent hover:text-foreground"
                        title="收起侧栏 (⌘B / Ctrl+B)"
                        aria-label="收起侧栏"
                    >
                        <PanelLeftClose className="h-4 w-4" />
                    </button>
                </div>
                {/* 分类区 */}
                <div className="border-b border-sidebar-border p-3">
                    <div className="mb-2 flex items-center justify-between">
                        <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                            分类
                        </span>
                        <button
                            type="button"
                            onClick={() => setShowCategoryForm(v => !v)}
                            className="inline-flex h-6 w-6 items-center justify-center rounded-sm text-muted-foreground hover:bg-sidebar-accent hover:text-violet-600 dark:hover:text-violet-400"
                            aria-label="新建分类"
                        >
                            <Plus className="h-3.5 w-3.5" />
                        </button>
                    </div>
                    {showCategoryForm && (
                        <div className="mb-2 flex gap-1">
                            <input
                                className={obsidianInput}
                                placeholder="分类名"
                                value={newCategoryName}
                                onChange={e => setNewCategoryName(e.target.value)}
                                onKeyDown={e => {
                                    if (e.key === 'Enter') handleCreateCategory()
                                    if (e.key === 'Escape') { setShowCategoryForm(false); setNewCategoryName('') }
                                }}
                                autoFocus
                            />
                            <button type="button" onClick={handleCreateCategory} className="rounded-sm p-1 text-violet-600 hover:bg-sidebar-accent dark:text-violet-400" aria-label="确认">
                                <Check className="h-3.5 w-3.5" />
                            </button>
                            <button type="button" onClick={() => { setShowCategoryForm(false); setNewCategoryName('') }} className="rounded-sm p-1 text-muted-foreground hover:bg-sidebar-accent" aria-label="取消">
                                <X className="h-3.5 w-3.5" />
                            </button>
                        </div>
                    )}
                    <div className="space-y-px">
                        {categories.map(cat => (
                            <div
                                key={cat.id}
                                className={`group ${sidebarRowBase} ${selectedCategoryId === cat.id ? sidebarRowActive : sidebarRowInactive}`}
                                onClick={() => {
                                    setSelectedCategoryId(cat.id)
                                    navigate('/note')
                                    void loadNotes(cat.id)
                                }}
                            >
                                <span className="min-w-0 flex-1 truncate">{cat.name}</span>
                                <button
                                    type="button"
                                    className="shrink-0 rounded-sm p-0.5 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100"
                                    onClick={e => { e.stopPropagation(); handleDeleteCategory(cat.id) }}
                                    aria-label="删除分类"
                                >
                                    <X className="h-3 w-3" />
                                </button>
                            </div>
                        ))}
                    </div>
                </div>

                {/* 标签区 */}
                <div className="border-b border-sidebar-border p-3">
                    <div className="mb-2 flex items-center justify-between">
                        <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                            标签
                        </span>
                        <button
                            type="button"
                            onClick={() => setShowTagForm(v => !v)}
                            className="inline-flex h-6 w-6 items-center justify-center rounded-sm text-muted-foreground hover:bg-sidebar-accent hover:text-violet-600 dark:hover:text-violet-400"
                            aria-label="新建标签"
                        >
                            <Plus className="h-3.5 w-3.5" />
                        </button>
                    </div>
                    {showTagForm && (
                        <div className="mb-2 flex gap-1">
                            <input
                                className={obsidianInput}
                                placeholder="标签名"
                                value={newTagNameInSidebar}
                                onChange={e => setNewTagNameInSidebar(e.target.value)}
                                onKeyDown={e => {
                                    if (e.key === 'Enter') handleCreateTagInSidebar()
                                    if (e.key === 'Escape') { setShowTagForm(false); setNewTagNameInSidebar('') }
                                }}
                                autoFocus
                            />
                            <button type="button" onClick={handleCreateTagInSidebar} className="rounded-sm p-1 text-violet-600 hover:bg-sidebar-accent dark:text-violet-400" aria-label="确认">
                                <Check className="h-3.5 w-3.5" />
                            </button>
                            <button type="button" onClick={() => { setShowTagForm(false); setNewTagNameInSidebar('') }} className="rounded-sm p-1 text-muted-foreground hover:bg-sidebar-accent" aria-label="取消">
                                <X className="h-3.5 w-3.5" />
                            </button>
                        </div>
                    )}
                    <div className="flex flex-wrap gap-1">
                        {allTags.map(tag => (
                            <div
                                key={tag.id}
                                className={`group flex cursor-pointer items-center gap-0.5 rounded-sm border px-1.5 py-0.5 text-[11px] transition-colors ${selectedTagIds.includes(tag.id)
                                    ? 'border-violet-500/50 bg-violet-500/10 text-foreground dark:border-violet-400/40 dark:bg-violet-400/10'
                                    : 'border-transparent bg-muted/50 text-muted-foreground hover:bg-sidebar-accent hover:text-foreground'
                                    }`}
                                onClick={() => handleToggleTagFilter(tag.id)}
                            >
                                <span className="max-w-28 truncate">{tag.name}</span>
                                <button
                                    type="button"
                                    className="shrink-0 rounded p-0.5 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100"
                                    onClick={e => { e.stopPropagation(); handleDeleteTag(tag.id) }}
                                    aria-label="删除标签"
                                >
                                    <X className="h-2.5 w-2.5" />
                                </button>
                            </div>
                        ))}
                        {allTags.length === 0 && (
                            <p className="py-1 text-[12px] text-muted-foreground">暂无标签</p>
                        )}
                    </div>
                    {selectedTagIds.length > 0 && (
                        <div className="mt-2 flex items-center justify-between">
                            <span className="text-[10px] text-muted-foreground">已选 {selectedTagIds.length} 个</span>
                            <button
                                type="button"
                                onClick={() => setSelectedTagIds([])}
                                className="text-[10px] text-violet-600 hover:underline dark:text-violet-400"
                            >
                                清除筛选
                            </button>
                        </div>
                    )}
                </div>

                {/* 笔记列表 */}
                <div className="flex-1 overflow-y-auto p-3">
                    <div className="mb-2 flex items-center justify-between">
                        <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                            笔记
                        </span>
                        <button
                            type="button"
                            onClick={() => setShowNoteForm(v => !v)}
                            className="inline-flex h-6 w-6 items-center justify-center rounded-sm text-muted-foreground hover:bg-sidebar-accent hover:text-violet-600 dark:hover:text-violet-400"
                            aria-label="新建笔记"
                        >
                            <Plus className="h-3.5 w-3.5" />
                        </button>
                    </div>
                    {showNoteForm && (
                        <div className="mb-2 flex gap-1">
                            <input
                                className={obsidianInput}
                                placeholder="笔记标题"
                                value={newNoteName}
                                onChange={e => setNewNoteName(e.target.value)}
                                onKeyDown={e => {
                                    if (e.key === 'Enter') handleCreateNote()
                                    if (e.key === 'Escape') { setShowNoteForm(false); setNewNoteName('') }
                                }}
                                autoFocus
                            />
                            <button type="button" onClick={handleCreateNote} className="rounded-sm p-1 text-violet-600 hover:bg-sidebar-accent dark:text-violet-400" aria-label="确认">
                                <Check className="h-3.5 w-3.5" />
                            </button>
                            <button type="button" onClick={() => { setShowNoteForm(false); setNewNoteName('') }} className="rounded-sm p-1 text-muted-foreground hover:bg-sidebar-accent" aria-label="取消">
                                <X className="h-3.5 w-3.5" />
                            </button>
                        </div>
                    )}
                    <div className="space-y-1">
                        {filteredNotes.map(note => (
                            <div key={note.id} className="relative">
                                {renamingNoteId === note.id ? (
                                    <div className="flex gap-1 py-1 pl-1">
                                        <input
                                            className={obsidianInput}
                                            value={renamingTitle}
                                            onChange={e => setRenamingTitle(e.target.value)}
                                            onKeyDown={e => {
                                                if (e.key === 'Enter') handleRenameNote(note, renamingTitle)
                                                if (e.key === 'Escape') { setRenamingNoteId(null); setRenamingTitle('') }
                                            }}
                                            autoFocus
                                        />
                                        <button type="button" onClick={() => handleRenameNote(note, renamingTitle)} className="rounded-sm p-1 text-violet-600 hover:bg-sidebar-accent dark:text-violet-400" aria-label="确认">
                                            <Check className="h-3.5 w-3.5" />
                                        </button>
                                        <button type="button" onClick={() => { setRenamingNoteId(null); setRenamingTitle('') }} className="rounded-sm p-1 text-muted-foreground hover:bg-sidebar-accent" aria-label="取消">
                                            <X className="h-3.5 w-3.5" />
                                        </button>
                                    </div>
                                ) : (
                                    <div
                                        className={`group ${sidebarRowBase} ${selectedNote?.id === note.id ? sidebarRowActive : sidebarRowInactive}`}
                                        onClick={() => handleSelectNote(note.id)}
                                    >
                                        <div className="min-w-0 flex-1">
                                            <div className="flex items-center justify-between gap-1">
                                                <span className="flex min-w-0 flex-1 items-center gap-1.5">
                                                    {selectedNote?.id === note.id && isNoteDirty && (
                                                        <span
                                                            className="h-2 w-2 shrink-0 rounded-full bg-violet-500 dark:bg-violet-400"
                                                            title="未保存的更改"
                                                            aria-hidden
                                                        />
                                                    )}
                                                    <span className="truncate">{note.title}</span>
                                                </span>
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <button
                                                            type="button"
                                                            className="rounded-sm p-0.5 text-muted-foreground opacity-0 hover:bg-sidebar-accent hover:text-foreground group-hover:opacity-100 data-[state=open]:opacity-100"
                                                            title="笔记操作"
                                                            aria-label="笔记操作"
                                                            onClick={e => e.stopPropagation()}
                                                        >
                                                            <MoreVertical className="h-3.5 w-3.5" />
                                                        </button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent
                                                        align="end"
                                                        className="w-52"
                                                        onClick={e => e.stopPropagation()}
                                                    >
                                                        <DropdownMenuItem
                                                            onSelect={() => {
                                                                setRenamingNoteId(note.id)
                                                                setRenamingTitle(note.title)
                                                            }}
                                                        >
                                                            <Pencil className="h-3.5 w-3.5" />
                                                            重命名
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem
                                                            onSelect={() =>
                                                                setEditingTagsNoteId(
                                                                    editingTagsNoteId === note.id ? null : note.id
                                                                )
                                                            }
                                                        >
                                                            <TagIcon className="h-3.5 w-3.5" />
                                                            编辑标签
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuSub>
                                                            <DropdownMenuSubTrigger>
                                                                移动到分类…
                                                            </DropdownMenuSubTrigger>
                                                            <DropdownMenuSubContent className="max-h-56 overflow-y-auto">
                                                                {categories.filter(c => c.id !== note.category.id)
                                                                    .length === 0 ? (
                                                                    <DropdownMenuItem disabled>
                                                                        暂无其它分类
                                                                    </DropdownMenuItem>
                                                                ) : (
                                                                    categories
                                                                        .filter(c => c.id !== note.category.id)
                                                                        .map(cat => (
                                                                            <DropdownMenuItem
                                                                                key={cat.id}
                                                                                onSelect={() => {
                                                                                    void handleMoveNoteToCategory(
                                                                                        note,
                                                                                        cat.id
                                                                                    )
                                                                                }}
                                                                            >
                                                                                {cat.name}
                                                                            </DropdownMenuItem>
                                                                        ))
                                                                )}
                                                            </DropdownMenuSubContent>
                                                        </DropdownMenuSub>
                                                        <DropdownMenuSub
                                                            onOpenChange={open => {
                                                                if (open) void loadVersionsForSubmenu(note.id)
                                                            }}
                                                        >
                                                            <DropdownMenuSubTrigger>
                                                                恢复历史版本…
                                                            </DropdownMenuSubTrigger>
                                                            <DropdownMenuSubContent className="max-h-56 w-[min(18rem,calc(100vw-2rem))] overflow-y-auto">
                                                                {versionsLoading &&
                                                                versionsSubNoteId === note.id ? (
                                                                    <DropdownMenuItem disabled>
                                                                        加载中…
                                                                    </DropdownMenuItem>
                                                                ) : versionsSubNoteId === note.id &&
                                                                  versionsList.length === 0 ? (
                                                                    <DropdownMenuItem disabled>
                                                                        暂无历史版本
                                                                    </DropdownMenuItem>
                                                                ) : (
                                                                    versionsSubNoteId === note.id &&
                                                                    versionsList.map(v => (
                                                                        <DropdownMenuItem
                                                                            key={v.id}
                                                                            className="flex items-center gap-1.5"
                                                                            onSelect={() => {
                                                                                void handleRestoreNoteVersion(
                                                                                    note,
                                                                                    v.id
                                                                                )
                                                                            }}
                                                                        >
                                                                            <span className="min-w-0 flex-1 truncate">
                                                                                {new Date(
                                                                                    v.created_at
                                                                                ).toLocaleString()}
                                                                            </span>
                                                                            <button
                                                                                type="button"
                                                                                className="inline-flex shrink-0 rounded-sm p-1 text-muted-foreground hover:bg-accent hover:text-foreground [&_svg]:pointer-events-auto"
                                                                                title="预览此版本"
                                                                                aria-label="预览此版本"
                                                                                onClick={e => {
                                                                                    e.stopPropagation()
                                                                                    void handlePreviewNoteVersion(
                                                                                        note,
                                                                                        v
                                                                                    )
                                                                                }}
                                                                                onPointerDown={e =>
                                                                                    e.stopPropagation()
                                                                                }
                                                                                onPointerUp={e =>
                                                                                    e.stopPropagation()
                                                                                }
                                                                            >
                                                                                <Eye className="h-3.5 w-3.5" />
                                                                            </button>
                                                                        </DropdownMenuItem>
                                                                    ))
                                                                )}
                                                            </DropdownMenuSubContent>
                                                        </DropdownMenuSub>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem
                                                            variant="destructive"
                                                            onSelect={() => {
                                                                if (
                                                                    confirm(
                                                                        '确定删除这篇笔记吗？'
                                                                    )
                                                                ) {
                                                                    void (async () => {
                                                                        try {
                                                                            await deleteNote(
                                                                                note.id
                                                                            )
                                                                            setNotes(prev =>
                                                                                prev.filter(
                                                                                    n =>
                                                                                        n.id !==
                                                                                        note.id
                                                                                )
                                                                            )
                                                                            if (
                                                                                selectedNote?.id ===
                                                                                note.id
                                                                            ) {
                                                                                navigate(
                                                                                    '/note'
                                                                                )
                                                                                setSelectedNote(
                                                                                    null
                                                                                )
                                                                            }
                                                                        } catch (err) {
                                                                            console.error(
                                                                                err
                                                                            )
                                                                        }
                                                                    })()
                                                                }
                                                            }}
                                                        >
                                                            <Trash2 className="h-3.5 w-3.5" />
                                                            删除笔记
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </div>
                                            {note.tags.length > 0 && (
                                                <div className="mt-1 flex flex-wrap gap-0.5">
                                                    {note.tags.slice(0, 2).map(tag => (
                                                        <span
                                                            key={tag.id}
                                                            className={`rounded-sm px-1 py-px text-[10px] ${selectedNote?.id === note.id
                                                                ? 'bg-violet-500/15 text-foreground dark:bg-violet-400/15'
                                                                : 'bg-muted/60 text-muted-foreground'
                                                                }`}
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
                                    </div>
                                )}
                                {editingTagsNoteId === note.id && (
                                    <div className="absolute left-0 right-0 z-20 mt-1 rounded-sm border border-sidebar-border bg-popover p-2 shadow-md">
                                        <div className="mb-2 flex flex-wrap gap-1">
                                            {allTags.map(tag => (
                                                <button
                                                    type="button"
                                                    key={tag.id}
                                                    onClick={() => handleToggleNoteTag(note, tag.id)}
                                                    className={`rounded-sm border px-1.5 py-0.5 text-[10px] transition-colors ${note.tags.some(t => t.id === tag.id)
                                                        ? 'border-violet-500/50 bg-violet-500/10 text-foreground dark:border-violet-400/40'
                                                        : 'border-transparent bg-muted/40 text-muted-foreground hover:bg-sidebar-accent'
                                                        }`}
                                                >
                                                    {tag.name}
                                                </button>
                                            ))}
                                        </div>
                                        <div className="flex gap-1">
                                            <input
                                                className={obsidianInput}
                                                placeholder="新建标签…"
                                                value={newTagName}
                                                onChange={e => setNewTagName(e.target.value)}
                                                onKeyDown={e => {
                                                    if (e.key === 'Enter') handleCreateAndAddTag(note)
                                                    if (e.key === 'Escape') setEditingTagsNoteId(null)
                                                }}
                                            />
                                            <button type="button" onClick={() => handleCreateAndAddTag(note)} className="rounded-sm p-1 text-violet-600 hover:bg-sidebar-accent dark:text-violet-400" aria-label="添加">
                                                <Plus className="h-3.5 w-3.5" />
                                            </button>
                                            <button type="button" onClick={() => setEditingTagsNoteId(null)} className="rounded-sm p-1 text-muted-foreground hover:bg-sidebar-accent" aria-label="关闭">
                                                <X className="h-3.5 w-3.5" />
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>
                        ))}
                        {filteredNotes.length === 0 && notes.length > 0 && (
                            <p className="py-2 text-[12px] text-muted-foreground">没有符合筛选条件的笔记</p>
                        )}
                        {notes.length === 0 && (
                            <p className="py-2 text-[12px] text-muted-foreground">暂无笔记</p>
                        )}
                    </div>
                </div>
            </div>

            {/* 右侧：画布区（扁平、无重阴影卡片） */}
            <div className="relative flex flex-1 flex-col overflow-hidden">
                {sidebarCollapsed && (
                    <button
                        type="button"
                        onClick={() => persistSidebarCollapsed(false)}
                        className="absolute left-0 top-12 z-50 inline-flex h-9 w-7 items-center justify-center rounded-r-md border border-l-0 border-border bg-muted/80 text-muted-foreground shadow-sm backdrop-blur-sm hover:bg-muted hover:text-foreground"
                        title="展开侧栏 (⌘B / Ctrl+B)"
                        aria-label="展开侧栏"
                    >
                        <PanelLeftOpen className="h-4 w-4" />
                    </button>
                )}
                {selectedNote ? (
                    <>
                        {!versionPreview && (
                            <div className="absolute right-4 top-4 z-50 flex items-center gap-1">
                                <button
                                    type="button"
                                    onClick={() => {
                                        if (viewMode === 'edit') {
                                            if (selectedNote && editingContent !== selectedNote.content_md) {
                                                handleUpdateNote()
                                            } else {
                                                setViewMode('view')
                                            }
                                        } else {
                                            startEdit()
                                        }
                                    }}
                                    className="inline-flex h-8 w-8 items-center justify-center rounded-sm border border-border bg-muted/30 text-muted-foreground hover:bg-muted hover:text-foreground"
                                    title={viewMode === 'edit' ? '预览' : '编辑'}
                                >
                                    {viewMode === 'edit' ? <Eye className="h-4 w-4" /> : <Pencil className="h-4 w-4" />}
                                </button>
                                <button
                                    type="button"
                                    onClick={handleProcessAI}
                                    disabled={aiProcessing}
                                    className="inline-flex h-8 w-8 items-center justify-center rounded-sm border border-border bg-muted/30 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
                                    title={aiProcessing ? `AI 处理中: ${aiStatus}` : 'AI 分析'}
                                >
                                    {aiProcessing ? (
                                        <Loader2 className="h-4 w-4 animate-spin text-violet-600 dark:text-violet-400" />
                                    ) : (
                                        <Sparkles className="h-4 w-4 text-violet-600 dark:text-violet-400" />
                                    )}
                                </button>
                            </div>
                        )}

                        <div className="flex h-full min-h-0 flex-col border-l border-border/60 bg-background">
                            <div className="border-b border-border/60 px-10 py-3 pr-24">
                                <div className="flex min-w-0 flex-col gap-2">
                                    <div className="flex min-w-0 items-center gap-2">
                                        {isNoteDirty && (
                                            <span
                                                className="h-2.5 w-2.5 shrink-0 rounded-full bg-violet-500 dark:bg-violet-400"
                                                title="有未保存的更改（⌘S / Ctrl+S 保存）"
                                            />
                                        )}
                                        <h1 className="min-w-0 flex-1 truncate text-lg font-semibold tracking-tight text-foreground">
                                            {selectedNote.title}
                                        </h1>
                                        {isNoteDirty && (
                                            <span className="hidden shrink-0 text-xs text-muted-foreground sm:inline">
                                                未保存
                                            </span>
                                        )}
                                    </div>
                                    {versionPreview && (
                                        <div className="flex flex-wrap items-center gap-2 rounded-sm border border-amber-500/25 bg-amber-500/10 px-2.5 py-1.5 text-[12px] text-foreground dark:border-amber-400/20 dark:bg-amber-400/10">
                                            <span className="min-w-0 text-muted-foreground">
                                                历史版本预览 ·{' '}
                                                {new Date(versionPreview.created_at).toLocaleString()}
                                            </span>
                                            <button
                                                type="button"
                                                onClick={() => setVersionPreview(null)}
                                                className="inline-flex shrink-0 items-center gap-1 rounded-sm px-2 py-0.5 text-[11px] font-medium text-foreground hover:bg-amber-500/20 dark:hover:bg-amber-400/15"
                                            >
                                                <X className="h-3 w-3" />
                                                关闭预览
                                            </button>
                                        </div>
                                    )}
                                </div>
                            </div>
                            <div className="flex min-h-0 flex-1 flex-col px-10 pb-6 pt-2">
                                {versionPreview ? (
                                    <div className="prose prose-sm max-w-208 min-h-0 flex-1 overflow-y-auto py-2 dark:prose-invert prose-a:text-violet-600 dark:prose-a:text-violet-400">
                                        <NoteMarkdownBody
                                            source={versionPreview.content_md}
                                            emptyHint="该历史版本没有正文"
                                        />
                                    </div>
                                ) : viewMode === 'edit' ? (
                                    <textarea
                                        className="min-h-0 w-full flex-1 resize-none border-0 bg-transparent py-2 font-mono text-[15px] leading-relaxed text-foreground outline-none selection:bg-violet-500/20 dark:selection:bg-violet-400/20"
                                        spellCheck={false}
                                        value={editingContent}
                                        onChange={e => setEditingContent(e.target.value)}
                                        autoFocus
                                    />
                                ) : (
                                    <div className="prose prose-sm max-w-208 min-h-0 flex-1 overflow-y-auto py-2 dark:prose-invert prose-a:text-violet-600 dark:prose-a:text-violet-400">
                                        {selectedNote.content_md ? (
                                            <NoteMarkdownBody
                                                source={editingContent || selectedNote.content_md}
                                                emptyHint="笔记为空，点击编辑按钮开始编辑"
                                            />
                                        ) : (
                                            <div className="text-sm italic text-muted-foreground">
                                                笔记为空，点击编辑按钮开始编辑
                                            </div>
                                        )}
                                    </div>
                                )}
                                <div
                                    className="mt-auto flex flex-wrap items-center gap-2 border-t border-border/40 pt-3 text-[11px] text-muted-foreground"
                                    onClick={handleClickOutside}
                                >
                                    <span>
                                        {versionPreview
                                            ? `该版本保存于 ${new Date(versionPreview.created_at).toLocaleString()}（当前笔记更新于 ${new Date(selectedNote.updated_at).toLocaleString()}）`
                                            : `更新于 ${new Date(selectedNote.updated_at).toLocaleString()}`}
                                    </span>
                                    {selectedNote.tags.length > 0 && (
                                        <div className="flex flex-wrap gap-1">
                                            {selectedNote.tags.map(tag => (
                                                <span
                                                    key={tag.id}
                                                    className="rounded-sm bg-muted/80 px-1.5 py-px text-[11px] text-muted-foreground"
                                                >
                                                    #{tag.name}
                                                </span>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    </>
                ) : (
                    <div className="flex h-full items-center justify-center border-l border-border/60 px-6 text-center text-sm text-muted-foreground">
                        {sidebarCollapsed
                            ? '点击左侧按钮或按 ⌘B / Ctrl+B 展开侧栏，选择或新建笔记'
                            : '选择一篇笔记，或点击侧栏 + 新建'}
                    </div>
                )}
            </div>
        </div>
    )
}
