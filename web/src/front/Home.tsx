import { useState, useEffect, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { front, PublicPostItem } from '../lib/api'

const categories = [
  { value: '', label: '全部' },
  { value: 'tech', label: '技术' },
  { value: 'life', label: '生活' },
  { value: 'treehole', label: '树洞' },
]

function formatDate(d: string) { return d.slice(0, 10) }

export default function Home() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [posts, setPosts] = useState<PublicPostItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)

  const page = parseInt(searchParams.get('page') || '1')
  const category = searchParams.get('category') || ''

  const fetchPosts = useCallback(() => {
    setLoading(true)
    front.posts({ page, per_page: 10, category })
      .then((res) => { setPosts(res.posts); setTotal(res.total) })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [page, category])

  useEffect(() => { fetchPosts() }, [fetchPosts])

  const setCategory = (cat: string) => {
    const p = new URLSearchParams(searchParams)
    if (cat) p.set('category', cat); else p.delete('category')
    p.delete('page')
    setSearchParams(p)
  }

  return (
    <div className="max-w-2xl mx-auto px-6 py-12">
      {/* Category filter */}
      <div className="flex items-center gap-1 mb-8">
        {categories.map(({ value, label }) => (
          <button
            key={value}
            onClick={() => setCategory(value)}
            className={`px-3 py-1 text-sm rounded-md transition-colors duration-150 ${
              category === value
                ? 'text-[var(--color-accent)]'
                : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Post list */}
      {loading ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">...</div>
      ) : posts.length === 0 ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">暂无文章</div>
      ) : (
        <div className="space-y-0">
          {posts.map((post, i) => (
            <div key={post.id}>
              <article className="py-5">
                <div className="flex items-baseline gap-4">
                  <time className="text-xs text-[var(--color-text-secondary)] whitespace-nowrap tabular-nums">
                    {post.published_at ? formatDate(post.published_at) : ''}
                  </time>
                  <div className="min-w-0">
                    <h2 className="text-base leading-snug">
                      <Link
                        to={`/post/${post.slug}`}
                        className="text-[var(--color-text)] hover:text-[var(--color-accent)] transition-colors duration-150"
                      >
                        {post.is_pinned && <span className="mr-1.5 text-[var(--color-accent)] text-xs">◆</span>}
                        {post.title || '(无标题)'}
                      </Link>
                    </h2>
                    <div className="flex items-center gap-2 mt-1">
                      <span className={`text-xs ${
                        post.category === 'tech' ? 'text-blue-400' :
                        post.category === 'life' ? 'text-green-400' :
                        'text-orange-400'
                      }`}>
                        {post.category === 'tech' ? '技术' : post.category === 'life' ? '生活' : '树洞'}
                      </span>
                      {post.summary && (
                        <p className="text-sm text-[var(--color-text-secondary)] truncate">
                          {post.summary.length > 80 ? post.summary.slice(0, 80) + '...' : post.summary}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              </article>
              {i < posts.length - 1 && <hr className="border-[var(--color-border)]" />}
            </div>
          ))}
        </div>
      )}

      {/* Pagination */}
      {total > 10 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          {Array.from({ length: Math.ceil(total / 10) }, (_, i) => i + 1).map((p) => (
            <button
              key={p}
              onClick={() => {
                const sp = new URLSearchParams(searchParams)
                sp.set('page', String(p))
                setSearchParams(sp)
              }}
              className={`w-8 h-8 rounded-md text-sm transition-colors duration-150 ${
                p === page
                  ? 'bg-[var(--color-accent)] text-white'
                  : 'bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
              }`}
            >
              {p}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
