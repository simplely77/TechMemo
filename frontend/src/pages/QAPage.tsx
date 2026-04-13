import { useState, useRef, useEffect, type MouseEvent } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  createSession,
  getSessions,
  sendMessageStream,
  getMessages,
  type CreateSessionResp,
  type ChatMessageResp,
  deleteSession,
  updateSession
} from "@/services/chatService"
import { Input } from "@/components/ui/input"
import { Pencil } from "lucide-react"
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

interface Message extends ChatMessageResp {
  localId: string
}

export default function QAPage() {
  const { sessionId: sessionIdParam } = useParams<{ sessionId?: string }>()
  const navigate = useNavigate()
  const [sessions, setSessions] = useState<CreateSessionResp[]>([])
  const [sessionsLoaded, setSessionsLoaded] = useState(false)
  const [currentSession, setCurrentSession] = useState<CreateSessionResp | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState("")
  const [loading, setLoading] = useState(false)
  const [renamingId, setRenamingId] = useState<number | null>(null)
  const [renameDraft, setRenameDraft] = useState("")
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messageInputRef = useRef<HTMLInputElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  useEffect(() => {
    loadSessions()
  }, [])

  useEffect(() => {
    if (currentSession) {
      loadMessages()
    }
  }, [currentSession])

  // 路由驱动：/qa 无会话；/qa/:sessionId 选中对应会话
  useEffect(() => {
    if (!sessionIdParam) {
      setCurrentSession(null)
      setMessages([])
      return
    }
    const id = Number(sessionIdParam)
    if (!Number.isFinite(id)) {
      navigate("/qa", { replace: true })
      return
    }
    const session = sessions.find((s) => s.id === id)
    if (session) {
      setCurrentSession(session)
    } else if (sessionsLoaded) {
      navigate("/qa", { replace: true })
    }
  }, [sessionIdParam, sessions, sessionsLoaded, navigate])

  // /qa 且无 sessionId 时，默认打开第一个会话
  useEffect(() => {
    if (sessionIdParam) return
    if (!sessionsLoaded || sessions.length === 0) return
    navigate(`/qa/${sessions[0].id}`, { replace: true })
  }, [sessionIdParam, sessionsLoaded, sessions, navigate])

  const loadSessions = async () => {
    try {
      const res = await getSessions(1, 20)
      setSessions(res.sessions || [])
    } catch (err) {
      console.error("Failed to load sessions", err)
    } finally {
      setSessionsLoaded(true)
    }
  }

  const loadMessages = async () => {
    if (!currentSession) return
    try {
      const res = await getMessages(currentSession.id, 1, 50)
      setMessages(res.messages.map(m => ({ ...m, localId: String(m.id) })))
    } catch (err) {
      console.error("Failed to load messages", err)
    }
  }

  const handleDeleteSession = async (id: number) => {
    try {
      await deleteSession(id)
      setSessions(prev => prev.filter(s => s.id !== id))
      if (currentSession?.id === id) {
        navigate("/qa", { replace: true })
      }
    } catch (err) {
      console.error("Failed to delete session", err)
    }
  }

  const handleCreateSession = async () => {
    try {
      // 默认标题「新会话 N」由服务端按已有「新会话 数字」会话递增生成
      const session = await createSession()
      setSessions(prev => [session, ...prev])
      navigate(`/qa/${session.id}`)
    } catch (err) {
      console.error("Failed to create session", err)
    }
  }

  const beginRename = (e: MouseEvent, session: CreateSessionResp) => {
    e.stopPropagation()
    setRenamingId(session.id)
    setRenameDraft(session.title || "")
  }

  const applyRename = async (sessionId: number) => {
    const title = renameDraft.trim()
    if (!title) return
    try {
      const updated = await updateSession(sessionId, { title })
      setSessions(prev =>
        prev.map(s => (s.id === sessionId ? { ...s, title: updated.title, updated_at: updated.updated_at } : s))
      )
      setCurrentSession(cur =>
        cur?.id === sessionId ? { ...cur, title: updated.title, updated_at: updated.updated_at } : cur
      )
      setRenamingId(null)
      setRenameDraft("")
    } catch (err) {
      console.error("Failed to rename session", err)
    }
  }

const handleSendMessage = async () => {
  if (!input.trim() || !currentSession) return

  const now = Date.now()

  const userMessage: Message = {
    id: now,
    session_id: currentSession.id,
    role: "user",
    content: input,
    created_at: new Date().toISOString(),
    localId: `user-${now}`
  }

  setMessages(prev => [...prev, userMessage])
  setInput("")
  setLoading(true)

  try {
    const response = await sendMessageStream(currentSession.id, { content: input })
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    let buffer = ""
    let assistantContent = ""

    const assistantMessage: Message = {
      id: now + 1,
      session_id: currentSession.id,
      role: "assistant",
      content: "",
      created_at: new Date().toISOString(),
      localId: `assistant-${now + 1}`
    }

    setMessages(prev => [...prev, assistantMessage])

    if (!reader) return

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      const events = buffer.split("\n\n")
      buffer = events.pop() || "" // 剩余不完整的留着

      for (const evt of events) {
        let eventType = "message"
        let dataLines: string[] = []

        const lines = evt.split("\n")

        for (const line of lines) {
          if (line.startsWith("event: ")) {
            eventType = line.slice(7).trim()
          } else if (line.startsWith("data: ")) {
            dataLines.push(line.slice(6))
          }
        }

        const data = dataLines.join("\n")

        if (eventType === "done") {
          break
        }

        if (eventType === "error") {
          console.error("SSE error:", data)
          continue
        }

        assistantContent += data

        setMessages(prev => {
          const updated = [...prev]
          updated[updated.length - 1] = {
            ...updated[updated.length - 1],
            content: assistantContent
          }
          return updated
        })
      }
    }
  } catch (err) {
    console.error("Failed to send message", err)
  } finally {
    setLoading(false)
    // 下一帧再聚焦，确保 loading 已解除、输入框不再 disabled
    setTimeout(() => messageInputRef.current?.focus(), 0)
  }
}

  return (
    <div className="h-full p-6 flex gap-6 overflow-hidden">
      <div className="max-w-6xl mx-auto w-full flex gap-6">
        {/* 左侧会话列表 */}
        <div className="w-64">
          <Card className="h-full">
            <CardHeader>
              <CardTitle className="text-xl flex">
                <p>会话列表</p>
                <Button size="sm" className="ml-auto" onClick={handleCreateSession}>
                  +
                </Button>
              </CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col h-full overflow-hidden">

              <ul className="space-y-1 flex-1 overflow-y-auto">
                {sessions.map(session => (
                  <li
                    key={session.id}
                    className={`p-2 border flex items-center gap-1 rounded cursor-pointer transition-all text-sm min-h-9 ${currentSession?.id === session.id
                      ? "bg-primary text-primary-foreground border-primary"
                      : "hover:bg-accent"
                      }`}
                    onClick={() => renamingId !== session.id && navigate(`/qa/${session.id}`)}
                  >
                    {renamingId === session.id ? (
                      <>
                        <Input
                          value={renameDraft}
                          onChange={e => setRenameDraft(e.target.value)}
                          className="flex-1 min-w-0 h-7 py-0 text-sm shadow-none bg-background text-black border-border focus-visible:ring-1"
                          autoFocus
                          onClick={e => e.stopPropagation()}
                          onMouseDown={e => e.stopPropagation()}
                          onKeyDown={e => {
                            if (e.key === "Enter") {
                              e.preventDefault()
                              const t = renameDraft.trim()
                              if (!t) {
                                setRenamingId(null)
                                setRenameDraft("")
                                return
                              }
                              void applyRename(session.id)
                            }
                            if (e.key === "Escape") {
                              e.preventDefault()
                              setRenamingId(null)
                              setRenameDraft("")
                            }
                          }}
                        />
                        <button
                          type="button"
                          className={`inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-colors ${
                            currentSession?.id === session.id
                              ? "text-primary-foreground/90 hover:bg-primary-foreground/15"
                              : "text-muted-foreground hover:bg-accent hover:text-foreground"
                          }`}
                          title="删除"
                          aria-label="删除"
                          onClick={e => {
                            e.stopPropagation()
                            handleDeleteSession(session.id)
                          }}
                        >
                          ✕
                        </button>
                      </>
                    ) : (
                      <>
                        <span className="truncate flex-1 min-w-0">
                          {session.title || `会话 ${session.id}`}
                        </span>
                        <button
                          type="button"
                          className={`inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-colors ${
                            currentSession?.id === session.id
                              ? "text-primary-foreground/90 hover:bg-primary-foreground/15 hover:text-primary-foreground"
                              : "text-muted-foreground hover:bg-accent hover:text-foreground"
                          }`}
                          title="重命名"
                          aria-label="重命名"
                          onClick={e => beginRename(e, session)}
                        >
                          <Pencil className="h-3.5 w-3.5" strokeWidth={2} />
                        </button>
                        <button
                          type="button"
                          className="text-xs opacity-60 hover:opacity-100 shrink-0"
                          title="删除"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleDeleteSession(session.id)
                          }}
                        >
                          ✕
                        </button>
                      </>
                    )}
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </div>

        {/* 右侧聊天区域 */}
        <div className="flex-1 flex flex-col">
          <h1 className="text-3xl font-bold mb-6">知识库问答</h1>

          <Card className="flex-1 flex flex-col mb-4">
            <CardContent className="flex-1 overflow-y-auto p-4 space-y-4">
              {messages.length === 0 ? (
                <div className="flex items-center justify-center h-full">
                  <p className="text-muted-foreground">开始提问，AI 将基于你的知识库回答</p>
                </div>
              ) : (
                messages.map(msg => (
                  <div
                    key={msg.localId}
                    className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}
                  >
                    <div
                      className="max-w-xs lg:max-w-md px-4 py-2 rounded-lg bg-muted text-muted-foreground"
                    >
                      <div className="text-sm prose prose-sm dark:prose-invert max-w-none">
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
                          {msg.content}
                        </ReactMarkdown>
                      </div>
                      <p className="text-xs opacity-70 mt-1">
                        {new Date(msg.created_at).toLocaleTimeString()}
                      </p>
                    </div>
                  </div>
                ))
              )}
              {loading && (
                <div className="flex justify-start">
                  <div className="bg-muted text-muted-foreground px-4 py-2 rounded-lg">
                    <p className="text-sm">AI 正在思考...</p>
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </CardContent>
          </Card>

          {/* 输入区域 */}
          <Card>
            <CardContent className="pt-4">
              <div className="flex gap-2">
                <input
                  ref={messageInputRef}
                  type="text"
                  placeholder="输入你的问题..."
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyPress={(e) => e.key === "Enter" && !loading && handleSendMessage()}
                  disabled={loading || !currentSession}
                  className="flex-1 px-3 py-2 border rounded"
                />
                <Button
                  onClick={handleSendMessage}
                  disabled={loading || !input.trim() || !currentSession}
                >
                  {loading ? "发送中..." : "发送"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
