import { useState, useEffect, useCallback, FormEvent } from 'react'
import { comments, CommentData } from '../lib/api'

function getInitial(name: string) {
  return name.charAt(0).toUpperCase()
}

function getColor(name: string) {
  const colors = ['#58a6ff', '#3fb950', '#d29922', '#f78166', '#bc8cff', '#f778ba']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

export default function CommentSection({ slug }: { slug: string }) {
  const [list, setList] = useState<CommentData[]>([])
  const [loading, setLoading] = useState(true)

  // Form
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [url, setUrl] = useState('')
  const [content, setContent] = useState('')
  const [remember, setRemember] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState('')
  const [captcha, setCaptcha] = useState({ a: 0, b: 0, answer: '' })

  const fetchComments = useCallback(() => {
    comments.getByPost(slug)
      .then((res) => setList(res.comments))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [slug])

  useEffect(() => {
    fetchComments()
    // Load saved info
    try {
      const saved = JSON.parse(localStorage.getItem('mo_comment_info') || '{}')
      if (saved.name) setName(saved.name)
      if (saved.email) setEmail(saved.email)
      if (saved.url) setUrl(saved.url)
      if (saved.remember) setRemember(true)
    } catch {}
    // Generate captcha
    setCaptcha((c) => ({ ...c, a: Math.floor(Math.random() * 9) + 1, b: Math.floor(Math.random() * 9) + 1 }))
  }, [fetchComments])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setMessage('')

    // Captcha check
    if (parseInt(captcha.answer) !== captcha.a + captcha.b) {
      setMessage('验证码错误')
      setCaptcha((c) => ({ ...c, a: Math.floor(Math.random() * 9) + 1, b: Math.floor(Math.random() * 9) + 1, answer: '' }))
      return
    }

    if (!name.trim() || !email.trim() || !content.trim()) {
      setMessage('请填写昵称、邮箱和评论内容')
      return
    }

    setSubmitting(true)
    try {
      await comments.create(slug, { author_name: name, author_email: email, author_url: url, content })
      setContent('')
      setCaptcha((c) => ({ ...c, a: Math.floor(Math.random() * 9) + 1, b: Math.floor(Math.random() * 9) + 1, answer: '' }))
      setMessage('评论已提交，待审核后显示')

      if (remember) {
        localStorage.setItem('mo_comment_info', JSON.stringify({ name, email, url, remember: true }))
      } else {
        localStorage.removeItem('mo_comment_info')
      }

      fetchComments()
    } catch (err: any) {
      setMessage(err.message || '提交失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h2 className="text-lg font-semibold text-[var(--color-text)] mb-6">评论 ({list.length})</h2>

      {/* Comment list */}
      {loading ? (
        <div className="text-sm text-[var(--color-text-secondary)]">加载中...</div>
      ) : list.length > 0 ? (
        <div className="space-y-5 mb-10">
          {list.map((c) => (
            <div key={c.id} className="flex gap-3">
              <div
                className="w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium text-white flex-shrink-0"
                style={{ backgroundColor: getColor(c.author_name) }}
              >
                {getInitial(c.author_name)}
              </div>
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-[var(--color-text)]">{c.author_name}</span>
                  <time className="text-xs text-[var(--color-text-secondary)]">
                    {new Date(c.created_at).toLocaleDateString('zh-CN')}
                  </time>
                </div>
                <p className="text-sm text-[var(--color-text)] mt-1 whitespace-pre-wrap">{c.content}</p>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="text-sm text-[var(--color-text-secondary)] mb-10">暂无评论</div>
      )}

      {/* Comment form */}
      <form onSubmit={handleSubmit} className="space-y-4">
        <h3 className="text-sm font-medium text-[var(--color-text)]">发表评论</h3>

        <div className="grid grid-cols-3 gap-3">
          <input
            type="text" value={name} onChange={(e) => setName(e.target.value)}
            placeholder="昵称 *" required
            className="px-3 py-2 rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)]"
          />
          <input
            type="email" value={email} onChange={(e) => setEmail(e.target.value)}
            placeholder="邮箱 *" required
            className="px-3 py-2 rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)]"
          />
          <input
            type="url" value={url} onChange={(e) => setUrl(e.target.value)}
            placeholder="网站 (可选)"
            className="px-3 py-2 rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)]"
          />
        </div>

        <textarea
          value={content} onChange={(e) => setContent(e.target.value)}
          placeholder="写下你的想法..." required rows={4}
          className="w-full px-3 py-2 rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)] resize-none"
        />

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" checked={remember} onChange={(e) => setRemember(e.target.checked)}
                className="accent-[var(--color-accent)]" />
              <span className="text-xs text-[var(--color-text-secondary)]">记住信息</span>
            </label>
            <div className="flex items-center gap-2">
              <span className="text-xs text-[var(--color-text-secondary)]">
                {captcha.a} + {captcha.b} = ?
              </span>
              <input
                type="text" value={captcha.answer} onChange={(e) => setCaptcha((c) => ({ ...c, answer: e.target.value }))}
                className="w-16 px-2 py-1 rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text)] text-sm text-center focus:border-[var(--color-accent)]"
              />
            </div>
          </div>
          <button
            type="submit" disabled={submitting}
            className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150"
          >
            {submitting ? '提交中...' : '发表评论'}
          </button>
        </div>

        {message && (
          <p className={`text-xs ${message.includes('错误') || message.includes('失败') ? 'text-red-400' : 'text-[var(--color-text-secondary)]'}`}>
            {message}
          </p>
        )}
      </form>
    </div>
  )
}
