import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { front, PublicPostItem } from '../lib/api'

export default function Treehole() {
  const [posts, setPosts] = useState<PublicPostItem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    front.posts({ category: 'treehole', per_page: 50 })
      .then((res) => setPosts(res.posts))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="max-w-xl mx-auto px-6 py-12">
      <h1 className="text-xl font-semibold text-[var(--color-text)] mb-8">树洞</h1>

      {loading ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">...</div>
      ) : posts.length === 0 ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">暂无碎碎念</div>
      ) : (
        <div className="relative">
          {/* Timeline line */}
          <div className="absolute left-0 top-2 bottom-2 w-px bg-[var(--color-border)]" />

          <div className="space-y-6">
            {posts.map((post) => (
              <div key={post.id} className="relative pl-6">
                {/* Dot */}
                <div className="absolute left-[-3px] top-2 w-[7px] h-[7px] rounded-full bg-[var(--color-border)]" />

                <time className="text-xs text-[var(--color-text-secondary)]">
                  {post.published_at ? post.published_at.slice(0, 16).replace('T', ' ') : ''}
                </time>

                <div className="mt-1 text-[15px] leading-relaxed text-[var(--color-text)] whitespace-pre-wrap">
                  {post.summary}
                </div>

                <Link
                  to={`/post/${post.slug}`}
                  className="text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-accent)] mt-1 inline-block"
                >
                  查看详情
                </Link>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
