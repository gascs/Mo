import { useState, useEffect } from 'react'
import { pages } from '../lib/api'

export default function Links() {
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    pages.get('links')
      .then((res) => setContent(res.content))
      .catch(() => setContent('<p class="text-[var(--color-text-secondary)]">此页面尚未编写内容。</p>'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="max-w-[720px] mx-auto px-6 py-12">
      <h1 className="text-xl font-semibold text-[var(--color-text)] mb-8">友情链接</h1>
      {loading ? (
        <div className="text-center text-[var(--color-text-secondary)] py-12">...</div>
      ) : (
        <div
          className="prose-custom text-[18px] leading-[1.8] text-[var(--color-text)]"
          dangerouslySetInnerHTML={{ __html: content }}
        />
      )}
    </div>
  )
}
