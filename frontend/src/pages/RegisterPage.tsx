import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { House } from 'lucide-react'
import { toast } from 'sonner'
import { register } from '../services/authService'
import { useAuthStore } from '../store/authStore'
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export default function RegisterPage() {
  const navigate = useNavigate()
  const token = useAuthStore((s) => s.token)

  const [form, setForm] = useState({ username: '', password: '', confirmPassword: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')

    if (form.password.length < 6) {
      setError('密码长度至少为 6 位')
      return
    }

    if (form.password !== form.confirmPassword) {
      setError('两次密码输入不一致')
      return
    }

    setLoading(true)
    try {
      await register({ username: form.username, password: form.password })
      navigate('/login', { state: { message: '注册成功，请登录' } })
    } catch (err: any) {
      setError(err.response?.data?.message || '注册失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const goHome = () => {
    if (token) navigate('/home')
    else toast.message('请先登录后再进入首页')
  }

  return (
    <div className="relative min-h-screen grid place-items-center">
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="absolute left-4 top-4 z-10 gap-1.5 text-muted-foreground"
        onClick={goHome}
      >
        <House className="h-4 w-4 shrink-0" aria-hidden />
        返回首页
      </Button>
      <Card className="w-full max-w-sm">
        <CardHeader className="flex flex-col items-center">
          <CardTitle className="text-2xl font-bold">注册</CardTitle>
          <CardAction className="w-full flex justify-end">
            <Button variant="link" onClick={() => navigate('/login')}>登录</Button>
          </CardAction>
        </CardHeader>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <CardContent>
            <div className="flex flex-col gap-4">
              <div className="grid gap-2">
                <Label htmlFor="username">昵称</Label>
                <Input
                  id="username"
                  type="text"
                  placeholder="请输入你的昵称"
                  required
                  value={form.username}
                  onChange={(e) => setForm({ ...form, username: e.target.value })}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="password">密码</Label>
                <Input id="password"
                  type="password"
                  placeholder="请输入你的密码"
                  required
                  value={form.password}
                  onChange={(e) => setForm({ ...form, password: e.target.value })}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="confirmPassword">确认密码</Label>
                <Input id="confirmPassword"
                  type="password"
                  placeholder="请再次输入密码确认"
                  required
                  value={form.confirmPassword}
                  onChange={(e) => setForm({ ...form, confirmPassword: e.target.value })}
                />
              </div>
            </div>
            {error && (
              <p className="text-red-500 text-sm mt-2">{error}</p>
            )}
          </CardContent>
          <CardFooter className="flex-col gap-2">
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "注册中..." : "注册"}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}
