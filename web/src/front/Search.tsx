import { useState, useEffect, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { front } from '../lib/api'

interface SearchItem {
  id: string; title: string; slug: string; summary: string; category: string; published_at: string | null;
}

export default function Search() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(searchParams.get('q') || '')
  const [results, setResults] = useState<SearchItem[]>([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)

  const doSearch = useCallback((q: string) => {
    if (!q.trim()) { setResults([]); setSearched(false); return }
    setLoading(true)
    setSearched(true)
    front.search(q)
      .then((res) => setResults(res.posts))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    const q = searchParams.get('q')
    if (q) { setQuery(q); doSearch(q) }
  }, [])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSearchParams({ q: query })
    doSearch(query)
  }

  const highlight = (text: string, q: string) => {
    if (!q.trim()) return text
    const re = new RegExp(`(${q.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
    return text.replace(re, '<mark class="bg-yellow-500/20 text-[var(--color-text)]">$1</mark>')
  }

  return (
    <div className="max-w-2xl mx-auto px-6 py-12">
      <form onSubmit={handleSubmit} className="mb-8">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="搜索文章..."
          className="w-full px-4 py-3 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text)] text-lg placeholder-[#484f58] focus:border-[var(--color-accent)] focus:shadow-[0_0_0_3px_rgba(88,166,255,0.3)] transition-all duration-150"
        />
      </form>

      {loading ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">搜索中...</div>
      ) : searched && results.length === 0 ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">没有找到匹配的文章</div>
      ) : (
        <div className="space-y-0">
          {results.map((item, i) => (
            <div key={item.id}>
              <article className="py-5">
                <h2 className="text-base font-medium">
                  <Link
                    to={`/post/${item.slug}`}
                    className="text-[var(--color-text)] hover:text-[var(--color-accent)] transition-colors duration-150"
                  >
                    <span dangerouslySetInnerHTML={{ __html: highlight(item.title || '(无标题)', query) }} />
                  </Link>
                </h2>
                <p
                  className="text-sm text-[var(--color-text-secondary)] mt-1"
                  dangerouslySetInnerHTML={{ __html: highlight(item.summary.slice(0, 150), query) }}
                />
                <div className="flex items-center gap-2 mt-1">
                  <span className="text-xs text-[var(--color-text-secondary)]">{item.published_at?.slice(0, 10)}</span>
                  <span className="text-xs text-[var(--color-text-secondary)]">{item.category}</span>
                </div>
              </article>
              {i < results.length - 1 && <hr className="border-[var(--color-border)]" />}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
