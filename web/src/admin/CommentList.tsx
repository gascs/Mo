import { useState, useEffect, useCallback } from 'react'
import { comments, CommentData } from '../lib/api'

const statuses = [
  { value: '', label: '全部' },
  { value: 'pending', label: '待审核' },
  { value: 'approved', label: '已通过' },
  { value: 'spam', label: '垃圾' },
  { value: 'trash', label: '已删除' },
]

export default function CommentList() {
  const [data, setData] = useState<{ comments: CommentData[]; total: number; page: number; per_page: number } | null>(null)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(() => {
    setLoading(true)
    comments.list({ page, per_page: 20, status })
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [page, status])

  useEffect(() => { fetchData() }, [fetchData])

  const handleStatus = async (id: string, newStatus: string) => {
    await comments.updateStatus(id, newStatus)
    fetchData()
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除此评论？')) return
    await comments.delete(id)
    fetchData()
  }

  const totalPages = data ? Math.ceil(data.total / data.per_page) : 0

  return (
    <div>
      <h1 className="text-2xl font-semibold text-[var(--color-text)] mb-6">评论管理</h1>

      {/* Status filter */}
      <div className="flex gap-2 mb-6">
        {statuses.map(({ value, label }) => (
          <button
            key={value}
            onClick={() => { setStatus(value); setPage(1) }}
            className={`px-3 py-1.5 rounded-md text-sm transition-colors duration-150 ${
              status === value
                ? 'bg-[var(--color-accent)] text-white'
                : 'bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">加载中...</div>
      ) : !data || data.comments.length === 0 ? (
        <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-8 text-center text-[var(--color-text-secondary)]">
          暂无评论
        </div>
      ) : (
        <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-[var(--color-border)] text-left">
                <th className="px-4 py-3 text-xs font-medium text-[var(--color-text-secondary)]">作者</th>
                <th className="px-4 py-3 text-xs font-medium text-[var(--color-text-secondary)]">内容</th>
                <th className="px-4 py-3 text-xs font-medium text-[var(--color-text-secondary)]">状态</th>
                <th className="px-4 py-3 text-xs font-medium text-[var(--color-text-secondary)]">时间</th>
                <th className="px-4 py-3 text-xs font-medium text-[var(--color-text-secondary)]">操作</th>
              </tr>
            </thead>
            <tbody>
              {data.comments.map((c) => (
                <tr key={c.id} className="border-b border-[var(--color-border)] last:border-0">
                  <td className="px-4 py-3">
                    <div className="text-sm text-[var(--color-text)]">{c.author_name}</div>
                    <div className="text-xs text-[var(--color-text-secondary)]">{c.author_email}</div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="text-sm text-[var(--color-text)] max-w-xs truncate">{c.content}</div>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      c.status === 'approved' ? 'bg-green-900/30 text-green-300' :
                      c.status === 'pending' ? 'bg-yellow-900/30 text-yellow-300' :
                      c.status === 'spam' ? 'bg-red-900/30 text-red-300' :
                      'bg-gray-800 text-gray-400'
                    }`}>
                      {c.status === 'approved' ? '已通过' :
                       c.status === 'pending' ? '待审核' :
                       c.status === 'spam' ? '垃圾' : '已删除'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-[var(--color-text-secondary)]">
                    {new Date(c.created_at).toLocaleDateString('zh-CN')}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-1 flex-wrap">
                      {c.status !== 'approved' && (
                        <button onClick={() => handleStatus(c.id, 'approved')}
                          className="text-xs text-green-400 hover:text-green-300 px-1">通过</button>
                      )}
                      {c.status !== 'spam' && (
                        <button onClick={() => handleStatus(c.id, 'spam')}
                          className="text-xs text-yellow-400 hover:text-yellow-300 px-1">垃圾</button>
                      )}
                      {c.status !== 'trash' && (
                        <button onClick={() => handleStatus(c.id, 'trash')}
                          className="text-xs text-gray-400 hover:text-gray-300 px-1">回收</button>
                      )}
                      <button onClick={() => handleDelete(c.id)}
                        className="text-xs text-red-400 hover:text-red-300 px-1">删除</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-6">
          {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
            <button key={p} onClick={() => setPage(p)}
              className={`w-8 h-8 rounded-md text-sm transition-colors duration-150 ${
                p === page ? 'bg-[var(--color-accent)] text-white' : 'bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
              }`}>
              {p}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
