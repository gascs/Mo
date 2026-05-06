import { useState, FormEvent, useEffect } from 'react'
import { api, setToken } from '../lib/api'

type Step = 1 | 2 | 3

export default function Setup() {
  const [step, setStep] = useState<Step>(1)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(true)

  // Form fields
  const [username, setUsername] = useState('admin')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [siteTitle, setSiteTitle] = useState('')
  const [siteSubtitle, setSiteSubtitle] = useState('')
  const [siteDesc, setSiteDesc] = useState('')

  useEffect(() => {
    api.setup.status().then((res: { setup_required: boolean }) => {
      if (!res.setup_required) {
        window.location.href = '/'
      }
    }).catch(() => {}).finally(() => setChecking(false))
  }, [])

  const handleStep1 = (e: FormEvent) => {
    e.preventDefault()
    setError('')

    if (password.length < 8) {
      setError('密码长度至少 8 位')
      return
    }
    if (password !== confirmPassword) {
      setError('两次密码不一致')
      return
    }
    // Password must contain both letters and digits
    if (!/[a-zA-Z]/.test(password) || !/[0-9]/.test(password)) {
      setError('密码必须包含字母和数字')
      return
    }

    setStep(2)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await api.setup.initialize({
        username,
        email,
        password,
        site_title: siteTitle,
        site_subtitle: siteSubtitle,
        site_desc: siteDesc,
      })
      setToken(res.access_token)
      setStep(3)
    } catch (err: any) {
      setError(err.message || '初始化失败')
    } finally {
      setLoading(false)
    }
  }

  if (checking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg)]">
        <p className="text-[var(--color-text-secondary)]">...</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg)]">
      <div className="w-full max-w-md mx-4">
        {/* Step 3: Complete */}
        {step === 3 ? (
          <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-8 text-center">
            <div className="w-12 h-12 rounded-full bg-[var(--color-accent)] flex items-center justify-center mx-auto mb-6">
              <svg className="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h1 className="text-xl font-semibold text-[var(--color-text)] mb-2">
              安装完成
            </h1>
            <p className="text-[var(--color-text-secondary)] mb-6">
              你的博客已经准备好了。请重启服务使配置生效。
            </p>
            <a
              href="/admin/login"
              className="inline-block px-6 py-2.5 rounded-md font-medium text-white bg-[var(--color-accent)] hover:opacity-90 transition-all duration-150"
            >
              前往登录
            </a>
          </div>
        ) : step === 2 ? (
          /* Step 2: Site info */
          <form
            onSubmit={handleSubmit}
            className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-8"
          >
            <h1 className="text-xl font-semibold text-[var(--color-text)] mb-2">
              站点信息
            </h1>
            <p className="text-sm text-[var(--color-text-secondary)] mb-8">
              第 2 步 / 2 — 填写你的博客基本信息
            </p>

            {error && (
              <div className="mb-6 p-3 rounded-md bg-red-900/20 border border-red-800 text-red-300 text-sm">
                {error}
              </div>
            )}

            <div className="space-y-5">
              <div>
                <label className="block text-sm text-[var(--color-text-secondary)] mb-1.5">
                  站点标题 *
                </label>
                <input
                  type="text"
                  value={siteTitle}
                  onChange={(e) => setSiteTitle(e.target.value)}
                  placeholder="我的博客"
                  required
                  className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
                />
              </div>

              <div>
                <label className="block text-sm text-[var(--color-text-secondary)] mb-1.5">
                  副标题（可选）
                </label>
                <input
                  type="text"
                  value={siteSubtitle}
                  onChange={(e) => setSiteSubtitle(e.target.value)}
                  placeholder="一个极简博客"
                  className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
                />
              </div>

              <div>
                <label className="block text-sm text-[var(--color-text-secondary)] mb-1.5">
                  站点描述（SEO）
                </label>
                <input
                  type="text"
                  value={siteDesc}
                  onChange={(e) => setSiteDesc(e.target.value)}
                  placeholder="用于搜索引擎展示的简短描述"
                  className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
                />
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setStep(1)}
                  className="flex-1 py-2.5 rounded-md font-medium text-[var(--color-text)] bg-[#21262d] border border-[var(--color-border)] hover:bg-[#30363d] transition-all duration-150"
                >
                  上一步
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="flex-1 py-2.5 rounded-md font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150"
                >
                  {loading ? '创建中...' : '完成安装'}
                </button>
              </div>
            </div>
          </form>
        ) : (
          /* Step 1: Admin account */
          <form
            onSubmit={handleStep1}
            className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-8"
          >
            <h1 className="text-xl font-semibold text-[var(--color-text)] mb-2">
              欢迎使用 Mo
            </h1>
            <p className="text-sm text-[var(--color-text-secondary)] mb-8">
              第 1 步 / 2 — 创建管理员账户
            </p>

            {error && (
              <div className="mb-6 p-3 rounded-md bg-red-900/20 border border-red-800 text-red-300 text-sm">
                {error}
              </div>
            )}

            <div className="space-y-5">
              <div>
                <label className="block text-sm text-[var(--color-text-secondary)] mb-1.5">
                  邮箱 *
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="admin@example.com"
                  required
                  autoComplete="email"
                  className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
                />
              </div>

              <div>
                <label className="block text-sm text-[var(--color-text-secondary)] mb-1.5">
                  用户名
                </label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="admin"
                  autoComplete="username"
                  className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
                />
              </div>

              <div>
                <label className="block text-sm text-[var(--color-text-secondary)] mb-1.5">
                  密码 *
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="至少8位，含字母和数字"
                  required
                  autoComplete="new-password"
                  className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
                />
                <div className="mt-1.5 flex gap-1">
                  {[1, 2, 3, 4].map((i) => (
                    <div
                      key={i}
                      className="h-1 flex-1 rounded-full transition-colors duration-150"
                      style={{
                        backgroundColor:
                          password.length >= 8 && /[a-zA-Z]/.test(password) && /[0-9]/.test(password)
                            ? i <= 4
                              ? '#3fb950'
                              : '#30363d'
                            : password.length >= i * 2
                            ? '#d29922'
                            : '#30363d',
                      }}
                    />
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-sm text-[var(--color-text-secondary)] mb-1.5">
                  确认密码 *
                </label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="再次输入密码"
                  required
                  autoComplete="new-password"
                  className="w-full px-4 py-2.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-colors duration-150"
                />
              </div>

              <button
                type="submit"
                className="w-full py-2.5 rounded-md font-medium text-white bg-[var(--color-accent)] hover:opacity-90 transition-all duration-150"
              >
                下一步
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}
