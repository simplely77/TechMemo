import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import HomePage from './pages/HomePage'
import NotePage from './pages/NotePage'
import KnowledgePage from './pages/KnowledgePage'
import SearchPage from './pages/SearchPage'
import QAPage from './pages/QAPage'
import StatsPage from './pages/StatsPage'
import MindmapPage from './pages/MindmapPage'

function ProtectedRoute() {
  const token = useAuthStore((s) => s.token)
  return token ? <Outlet />:<Navigate to="/login" replace/>
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/home" replace />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route element={<ProtectedRoute/>}>
          <Route path="/home" element={<HomePage/>} />
          <Route path="/notes" element={<NotePage/>} />
          <Route path="/knowledge" element={<KnowledgePage/>} />
          <Route path="/search" element={<SearchPage/>} />
          <Route path="/qa" element={<QAPage/>} />
          <Route path="/stats" element={<StatsPage/>} />
          <Route path="/mindmap" element={<MindmapPage/>} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
