import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { front } from '../lib/api'

interface ArchiveItem {
  id: string; title: string; slug: string; category: string; published_at: string | null;
}

export default function Archive() {
  const [groups, setGroups] = useState<{ year: string; posts: ArchiveItem[] }[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    front.archive()
      .then((res) => setGroups(res.archive))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="max-w-2xl mx-auto px-6 py-12">
      <h1 className="text-xl font-semibold text-[var(--color-text)] mb-8">归档</h1>

      {loading ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">...</div>
      ) : groups.length === 0 ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">暂无文章</div>
      ) : (
        <div className="space-y-10">
          {groups.map(({ year, posts }) => (
            <div key={year} className="flex gap-8">
              <div className="text-3xl font-bold text-[var(--color-border)] flex-shrink-0 w-20">
                {year}
              </div>
              <div className="flex-1 space-y-3">
                {posts.map((post) => (
                  <div key={post.id} className="flex items-baseline gap-4">
                    <time className="text-xs text-[var(--color-text-secondary)] whitespace-nowrap tabular-nums">
                      {post.published_at ? post.published_at.slice(0, 10) : ''}
                    </time>
                    <Link
                      to={`/post/${post.slug}`}
                      className="text-[var(--color-text)] hover:text-[var(--color-accent)] transition-colors duration-150 text-sm"
                    >
                      {post.title || '(无标题)'}
                    </Link>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
