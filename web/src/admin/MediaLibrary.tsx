import { useState, useEffect, useCallback } from 'react'
import { api } from '../lib/api'

interface MediaItem {
  id: string; file_name: string; original_name: string; file_path: string;
  file_size: number; mime_type: string; created_at: string;
}

export default function MediaLibrary() {
  const [items, setItems] = useState<MediaItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [uploading, setUploading] = useState(false)
  const [copied, setCopied] = useState<string | null>(null)

  const fetchMedia = useCallback(() => {
    api.media.list({ page, per_page: 20 }).then((res: any) => {
      setItems(res.media)
      setTotal(res.total)
    })
  }, [page])

  useEffect(() => { fetchMedia() }, [fetchMedia])

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (!files?.length) return
    setUploading(true)
    for (const file of Array.from(files)) {
      try { await api.media.upload(file) } catch (err) { console.error(err) }
    }
    setUploading(false)
    fetchMedia()
  }

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    const files = e.dataTransfer.files
    if (!files.length) return
    setUploading(true)
    for (const file of Array.from(files)) {
      try { await api.media.upload(file) } catch (err) { console.error(err) }
    }
    setUploading(false)
    fetchMedia()
  }

  const handleCopy = (filePath: string) => {
    const md = `![](/${filePath})`
    navigator.clipboard.writeText(md).then(() => {
      setCopied(filePath)
      setTimeout(() => setCopied(null), 2000)
    })
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除？')) return
    await api.media.delete(id)
    fetchMedia()
  }

  const totalPages = Math.ceil(total / 20)

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-[var(--color-text)]">媒体库</h1>
        <label className={`px-4 py-2 rounded-md text-sm font-medium text-white cursor-pointer transition-all duration-150 ${
          uploading ? 'opacity-50' : 'bg-[var(--color-accent)] hover:opacity-90'
        }`}>
          {uploading ? '上传中...' : '上传图片'}
          <input type="file" accept="image/*" multiple onChange={handleUpload} className="hidden" />
        </label>
      </div>

      <div
        onDrop={handleDrop}
        onDragOver={(e) => e.preventDefault()}
        className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-8"
      >
        {items.length === 0 ? (
          <div className="text-center text-[var(--color-text-secondary)] py-12">
            <p className="mb-2">拖拽图片到此处上传</p>
            <p className="text-xs">支持 JPG、PNG、GIF、WebP、SVG，最大 10MB</p>
          </div>
        ) : (
          <div className="grid grid-cols-4 gap-4">
            {items.map((m) => (
              <div
                key={m.id}
                className="group relative bg-[#0d1117] rounded-lg overflow-hidden border border-[var(--color-border)] hover:border-[var(--color-accent)] transition-colors"
              >
                <div className="aspect-square flex items-center justify-center bg-[#161b22]">
                  <img
                    src={`/uploads/${m.file_path}`}
                    alt={m.original_name}
                    className="max-w-full max-h-full object-contain"
                    loading="lazy"
                  />
                </div>
                <div className="p-2">
                  <div className="text-xs text-[var(--color-text-secondary)] truncate" title={m.original_name}>
                    {m.original_name}
                  </div>
                  <div className="flex items-center gap-2 mt-1">
                    <button
                      onClick={() => handleCopy(m.file_path)}
                      className="text-xs text-[var(--color-accent)] hover:underline"
                    >
                      {copied === m.file_path ? '已复制!' : '复制链接'}
                    </button>
                    <button
                      onClick={() => handleDelete(m.id)}
                      className="text-xs text-red-400 hover:text-red-300"
                    >
                      删除
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-6">
          {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
            <button
              key={p}
              onClick={() => setPage(p)}
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
