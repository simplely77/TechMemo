import { useState, useRef, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  createSession,
  getSessions,
  sendMessageStream,
  getMessages,
  type CreateSessionResp,
  type ChatMessageResp,
  deleteSession
} from "@/services/chatService"
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
  const [sessions, setSessions] = useState<CreateSessionResp[]>([])
  const [currentSession, setCurrentSession] = useState<CreateSessionResp | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState("")
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

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

  const loadSessions = async () => {
    try {
      const res = await getSessions(1, 20)
      setSessions(res.sessions || [])
      if (res.sessions && res.sessions.length > 0) {
        setCurrentSession(res.sessions[0])
      }
    } catch (err) {
      console.error("Failed to load sessions", err)
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
      if (currentSession?.id === id) {
        setCurrentSession(null)
      }
      setSessions(prev => prev.filter(s => s.id !== id))
    } catch (err) {
      console.error("Failed to delete session", err)
    }
  }

  const handleCreateSession = async () => {
    try {
      const session = await createSession()
      setSessions([session, ...sessions])
      setCurrentSession(session)
      setMessages([])
    } catch (err) {
      console.error("Failed to create session", err)
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
                    className={`p-2 border flex items-center rounded cursor-pointer transition-all text-sm truncate ${currentSession?.id === session.id
                      ? "bg-primary text-primary-foreground border-primary"
                      : "hover:bg-accent"
                      }`}
                    onClick={() => setCurrentSession(session)}
                  >
                    {/* 左边标题 */}
                    <span className="truncate flex-1">
                      {session.title || `会话 ${session.id}`}
                    </span>

                    {/* 右边删除按钮 */}
                    <button
                      className="text-xs opacity-60 hover:opacity-100 ml-auto mr-2"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteSession(session.id);
                      }}
                    >
                      ✕
                    </button>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </div>

        {/* 右侧聊天区域 */}
        <div className="flex-1 flex flex-col">
          <h1 className="text-3xl font-bold mb-6">智能问答</h1>

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
