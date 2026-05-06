import { useState, FormEvent, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, setToken } from '../lib/api'

export default function Login() {
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [remember, setRemember] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const saved = localStorage.getItem('mo_login')
    if (saved) {
      try {
        const { login: l, remember: r } = JSON.parse(saved)
        if (r) {
          setLogin(l)
          setRemember(true)
        }
      } catch {}
    }
  }, [])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')

    if (!login.trim() || !password) {
      setError('请输入用户名/邮箱和密码')
      return
    }

    setLoading(true)
    try {
      const res = await api.auth.login(login, password)
      setToken(res.access_token)

      if (remember) {
        localStorage.setItem('mo_login', JSON.stringify({ login, remember: true }))
      } else {
        localStorage.removeItem('mo_login')
      }

      navigate('/admin')
    } catch (err: any) {
      setError(err.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg)]">
      <div className="w-full max-w-sm mx-4">
        <form
          onSubmit={handleSubmit}
          className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-8"
        >
          <h1 className="text-xl font-semibold text-center mb-8 text-[var(--color-text)]">
            登录
          </h1>

          {error && (
            <div className="mb-6 p-3 rounded-md bg-red-900/20 border border-red-800 text-red-300 text-sm">
              {error}
            </div>
          )}

          <div className="space-y-5">
            <div>
              <input
                type="text"
                value={login}
                onChange={(e) => setLogin(e.target.value)}
                placeholder="邮箱或用户名"
                autoComplete="username"
                className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
              />
            </div>

            <div>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="密码"
                autoComplete="current-password"
                className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
              />
            </div>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={remember}
                onChange={(e) => setRemember(e.target.checked)}
                className="w-4 h-4 rounded border-[var(--color-border)] bg-[var(--color-bg)] accent-[var(--color-accent)]"
              />
              <span className="text-sm text-[var(--color-text-secondary)]">
                记住我
              </span>
            </label>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-2.5 rounded-md font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150"
            >
              {loading ? '登录中...' : '登录'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
