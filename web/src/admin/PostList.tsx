import { useState, useEffect } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api, PostData } from '../lib/api'

const categories = [
  { value: '', label: '全部分类' },
  { value: 'tech', label: '技术' },
  { value: 'life', label: '生活' },
  { value: 'treehole', label: '树洞' },
]

const statuses = [
  { value: '', label: '全部状态' },
  { value: 'published', label: '已发布' },
  { value: 'draft', label: '草稿' },
  { value: 'trashed', label: '回收站' },
]

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

export default function PostList() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [data, setData] = useState<{ posts: PostData[]; total: number; page: number; per_page: number } | null>(null)
  const [loading, setLoading] = useState(true)

  const page = parseInt(searchParams.get('page') || '1')
  const category = searchParams.get('category') || ''
  const status = searchParams.get('status') || ''
  const search = searchParams.get('search') || ''

  const fetchPosts = () => {
    setLoading(true)
    api.posts.list({ page, per_page: 10, category, status, search })
      .then(setData)
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchPosts() }, [page, category, status, search])

  const setFilter = (key: string, value: string) => {
    const p = new URLSearchParams(searchParams)
    if (value) p.set(key, value); else p.delete(key)
    if (key !== 'page') p.delete('page')
    setSearchParams(p)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除？文章将移入回收站。')) return
    await api.posts.delete(id)
    fetchPosts()
  }

  const handlePublish = async (id: string, publish: boolean) => {
    await api.posts.publish(id, publish)
    fetchPosts()
  }

  const totalPages = data ? Math.ceil(data.total / data.per_page) : 0

  return (
    <div>
      {/* Toolbar */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-[var(--color-text)]">文章</h1>
        <Link
          to="/admin/posts/new"
          className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 transition-all duration-150"
        >
          写文章
        </Link>
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-6 flex-wrap">
        <input
          type="text"
          placeholder="搜索..."
          value={search}
          onChange={(e) => setFilter('search', e.target.value)}
          className="px-3 py-1.5 rounded-md bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)] w-48"
        />
        {categories.map(({ value, label }) => (
          <button
            key={value}
            onClick={() => setFilter('category', value)}
            className={`px-3 py-1.5 rounded-md text-sm transition-colors duration-150 ${
              category === value
                ? 'bg-[var(--color-accent)] text-white'
                : 'bg-[#21262d] text-[var(--color-text-secondary)] border border-[var(--color-border)] hover:text-[var(--color-text)]'
            }`}
          >
            {label}
          </button>
        ))}
        <div className="w-px bg-[var(--color-border)]" />
        {statuses.map(({ value, label }) => (
          <button
            key={value}
            onClick={() => setFilter('status', value)}
            className={`px-3 py-1.5 rounded-md text-sm transition-colors duration-150 ${
              status === value
                ? 'bg-[var(--color-accent)] text-white'
                : 'bg-[#21262d] text-[var(--color-text-secondary)] border border-[var(--color-border)] hover:text-[var(--color-text)]'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-[var(--color-text-secondary)]">加载中...</div>
        ) : !data || data.posts.length === 0 ? (
          <div className="p-8 text-center text-[var(--color-text-secondary)]">
            暂无文章
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-[var(--color-border)] text-left">
                <th className="px-5 py-3 text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">标题</th>
                <th className="px-5 py-3 text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">分类</th>
                <th className="px-5 py-3 text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">状态</th>
                <th className="px-5 py-3 text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">日期</th>
                <th className="px-5 py-3 text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody>
              {data.posts.map((post) => (
                <tr key={post.id} className="border-b border-[var(--color-border)] last:border-0 hover:bg-[#1c2129] transition-colors">
                  <td className="px-5 py-3">
                    <Link to={`/admin/posts/${post.id}/edit`} className="text-[var(--color-text)] hover:text-[var(--color-accent)] transition-colors">
                      {post.title || '(无标题)'}
                    </Link>
                    {post.is_pinned && (
                      <span className="ml-2 text-xs text-[var(--color-accent)]">置顶</span>
                    )}
                  </td>
                  <td className="px-5 py-3">
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      post.category === 'tech' ? 'bg-blue-900/30 text-blue-300' :
                      post.category === 'life' ? 'bg-green-900/30 text-green-300' :
                      'bg-orange-900/30 text-orange-300'
                    }`}>
                      {post.category === 'tech' ? '技术' : post.category === 'life' ? '生活' : '树洞'}
                    </span>
                  </td>
                  <td className="px-5 py-3">
                    <span className={`text-xs ${post.deleted_at ? 'text-red-400' : post.is_draft ? 'text-yellow-400' : 'text-green-400'}`}>
                      {post.deleted_at ? '已删除' : post.is_draft ? '草稿' : '已发布'}
                    </span>
                  </td>
                  <td className="px-5 py-3 text-sm text-[var(--color-text-secondary)]">
                    {formatDate(post.created_at)}
                  </td>
                  <td className="px-5 py-3">
                    <div className="flex gap-2">
                      <Link
                        to={`/admin/posts/${post.id}/edit`}
                        className="text-xs text-[var(--color-accent)] hover:underline"
                      >
                        编辑
                      </Link>
                      {!post.deleted_at && (
                        <>
                          <button
                            onClick={() => handlePublish(post.id, post.is_draft)}
                            className="text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)]"
                          >
                            {post.is_draft ? '发布' : '撤回'}
                          </button>
                          <button
                            onClick={() => handleDelete(post.id)}
                            className="text-xs text-red-400 hover:text-red-300"
                          >
                            删除
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-6">
          {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
            <button
              key={p}
              onClick={() => setFilter('page', String(p))}
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
