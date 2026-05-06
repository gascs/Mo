import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'

interface Stats {
  total_posts: number
  published_posts: number
  draft_posts: number
  treehole_posts: number
}

const statCards = [
  { key: 'total_posts' as const, label: '总文章', color: '#58a6ff' },
  { key: 'published_posts' as const, label: '已发布', color: '#3fb950' },
  { key: 'draft_posts' as const, label: '草稿', color: '#d29922' },
  { key: 'treehole_posts' as const, label: '树洞', color: '#f78166' },
]

export default function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null)

  useEffect(() => {
    api.posts.dashboard().then(setStats).catch(console.error)
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-2xl font-semibold text-[var(--color-text)]">仪表盘</h1>
        <Link
          to="/admin/posts/new"
          className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 transition-all duration-150"
        >
          写文章
        </Link>
      </div>

      <div className="grid grid-cols-4 gap-4 mb-8">
        {statCards.map(({ key, label, color }) => (
          <div
            key={key}
            className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-5"
          >
            <div className="text-3xl font-bold mb-1" style={{ color }}>
              {stats ? stats[key] : '-'}
            </div>
            <div className="text-sm text-[var(--color-text-secondary)]">{label}</div>
          </div>
        ))}
      </div>

      <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
        <h2 className="text-sm font-semibold text-[var(--color-text-secondary)] uppercase tracking-wider mb-4">
          快速操作
        </h2>
        <div className="flex gap-3">
          <Link
            to="/admin/posts/new"
            className="px-4 py-2 rounded-md text-sm bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[#30363d] transition-all duration-150"
          >
            新建文章
          </Link>
          <Link
            to="/admin/posts/new?category=treehole"
            className="px-4 py-2 rounded-md text-sm bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[#30363d] transition-all duration-150"
          >
            写树洞
          </Link>
          <Link
            to="/admin/media"
            className="px-4 py-2 rounded-md text-sm bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[#30363d] transition-all duration-150"
          >
            媒体库
          </Link>
        </div>
      </div>
    </div>
  )
}
