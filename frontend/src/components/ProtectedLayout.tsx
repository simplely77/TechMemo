import { Link, Outlet, useLocation } from 'react-router-dom'
import { House } from 'lucide-react'

/**
 * 登录后各功能页的公共壳层：非首页时顶部提供「返回首页」。
 */
export default function ProtectedLayout() {
    const { pathname } = useLocation()
    const onHome = pathname === '/home'

    return (
        <div className="flex h-dvh max-h-dvh w-full flex-col overflow-hidden bg-background">
            {!onHome && (
                <header className="flex h-11 shrink-0 items-center border-b border-border bg-background/95 px-3 backdrop-blur supports-backdrop-filter:bg-background/80">
                    <Link
                        to="/home"
                        className="inline-flex items-center gap-1.5 rounded-md px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    >
                        <House className="h-4 w-4 shrink-0" aria-hidden />
                        返回首页
                    </Link>
                </header>
            )}
            <div className="flex min-h-0 flex-1 flex-col overflow-auto">
                <Outlet />
            </div>
        </div>
    )
}
