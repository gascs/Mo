import { useState, useEffect, useRef } from 'react'
import { useParams, Link } from 'react-router-dom'
import { front, PostData } from '../lib/api'
import CommentSection from './CommentSection'

function readingTime(html: string): number {
  const text = html.replace(/<[^>]*>/g, '')
  return Math.max(1, Math.ceil(text.length / 200))
}

export default function PostDetail() {
  const { slug } = useParams<{ slug: string }>()
  const [post, setPost] = useState<PostData | null>(null)
  const [loading, setLoading] = useState(true)
  const [showTop, setShowTop] = useState(false)
  const contentRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!slug) return
    front.postBySlug(slug).then(setPost).catch(() => {}).finally(() => setLoading(false))
  }, [slug])

  useEffect(() => {
    const onScroll = () => setShowTop(window.scrollY > 300)
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  if (loading) {
    return <div className="max-w-2xl mx-auto px-6 py-12 text-center text-[var(--color-text-secondary)]">...</div>
  }

  if (!post) {
    return (
      <div className="max-w-2xl mx-auto px-6 py-12 text-center">
        <p className="text-[var(--color-text-secondary)] mb-4">文章不存在</p>
        <Link to="/" className="text-[var(--color-accent)]">← 返回首页</Link>
      </div>
    )
  }

  const tags: string[] = (() => { try { return JSON.parse(post.tags) } catch { return [] } })()

  return (
    <div className="relative">
      <article className="max-w-[720px] mx-auto px-6 py-12">
        {/* Header */}
        <header className="mb-10">
          <h1 className="text-[2em] font-bold tracking-[-0.02em] leading-tight text-[var(--color-text)]">
            {post.title}
          </h1>
          <div className="flex items-center gap-3 mt-4 text-sm text-[var(--color-text-secondary)]">
            {post.published_at && <time>{new Date(post.published_at).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })}</time>}
            <span>·</span>
            <span>{post.category === 'tech' ? '技术' : post.category === 'life' ? '生活' : '树洞'}</span>
            <span>·</span>
            <span>约 {readingTime(post.content_html)} 分钟</span>
          </div>
          {tags.length > 0 && (
            <div className="flex gap-2 mt-3">
              {tags.map((t) => (
                <span key={t} className="text-xs px-2 py-0.5 rounded bg-[#21262d] text-[var(--color-text-secondary)]">{t}</span>
              ))}
            </div>
          )}
        </header>

        {/* Content */}
        <div
          ref={contentRef}
          className="prose-custom text-[18px] leading-[1.8] text-[var(--color-text)]"
          style={{
            wordBreak: 'break-word',
          }}
          dangerouslySetInnerHTML={{ __html: post.content_html }}
        />

        {/* Back navigation */}
        <div className="mt-12 pt-6 border-t border-[var(--color-border)]">
          <Link to="/" className="text-sm text-[var(--color-accent)] hover:underline">← 返回首页</Link>
        </div>
      </article>

      {/* Comments */}
      <div className="max-w-[720px] mx-auto px-6 pb-12">
        <CommentSection slug={slug!} />
      </div>

      {/* Back to top */}
      {showTop && (
        <button
          onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
          className="fixed bottom-8 right-8 w-10 h-10 rounded-full bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text)] flex items-center justify-center transition-all duration-150 opacity-60 hover:opacity-100"
        >
          ↑
        </button>
      )}
    </div>
  )
}
