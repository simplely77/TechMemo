import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import HomePage from './pages/HomePage'
import NotePage from './pages/NotePage'

function ProtectedRoute() {
  const token = useAuthStore((s) => s.token)
  return token ? <Outlet />:<Navigate to="/login" replace/>
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route element={<ProtectedRoute/>}>
          <Route path="/home" element={<HomePage/>} />
          <Route path="/note" element={<NotePage/>} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
