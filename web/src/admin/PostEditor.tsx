import { useEffect, useRef, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { markdown, markdownLanguage } from '@codemirror/lang-markdown'
import { oneDark } from '@codemirror/theme-one-dark'
import { syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'
import { marked } from 'marked'
import { api, PostData, setToken, getToken } from '../lib/api'

const categories = [
  { value: 'tech', label: '技术' },
  { value: 'life', label: '生活' },
  { value: 'treehole', label: '树洞' },
]

export default function PostEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const editorRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const [post, setPost] = useState<PostData | null>(null)
  const [title, setTitle] = useState('')
  const [category, setCategory] = useState('tech')
  const [tagsInput, setTagsInput] = useState('')
  const [slug, setSlug] = useState('')
  const [summary, setSummary] = useState('')
  const [isDraft, setIsDraft] = useState(true)
  const [isPrivate, setIsPrivate] = useState(false)
  const [isPinned, setIsPinned] = useState(false)
  const [saving, setSaving] = useState(false)
  const [preview, setPreview] = useState('')
  const [showPreview, setShowPreview] = useState(true)
  const saveTimer = useRef<ReturnType<typeof setTimeout>>()

  // Load existing post
  useEffect(() => {
    if (!id) return
    api.posts.get(id).then((p) => {
      setPost(p)
      setTitle(p.title)
      setCategory(p.category)
      setSlug(p.slug)
      setSummary(p.summary)
      setIsDraft(p.is_draft)
      setIsPrivate(p.is_private)
      setIsPinned(p.is_pinned)
      try {
        setTagsInput(JSON.parse(p.tags).join(', '))
      } catch { setTagsInput('') }
    })
  }, [id])

  // Init CodeMirror
  useEffect(() => {
    if (!editorRef.current) return
    const extensions = [
      lineNumbers(),
      highlightActiveLine(),
      history(),
      markdown({ base: markdownLanguage }),
      oneDark,
      syntaxHighlighting(defaultHighlightStyle),
      keymap.of([...defaultKeymap, ...historyKeymap]),
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const content = update.state.doc.toString()
          setPreview(marked.parse(content) as string)
          // Auto-save
          if (saveTimer.current) clearTimeout(saveTimer.current)
          saveTimer.current = setTimeout(() => autoSave(content), 2000)
        }
      }),
    ]

    const state = EditorState.create({
      doc: post?.content || '',
      extensions,
    })

    const view = new EditorView({ state, parent: editorRef.current })
    viewRef.current = view

    // Set initial preview
    if (post?.content) {
      setPreview(marked.parse(post.content) as string)
    }

    return () => view.destroy()
  }, [post?.id])

  const getContent = useCallback(() => {
    return viewRef.current?.state.doc.toString() || ''
  }, [])

  const autoSave = async (content: string) => {
    if (!title.trim() && !content.trim()) return
    setSaving(true)
    try {
      const tags = tagsInput.split(',').map((t) => t.trim()).filter(Boolean)
      const payload = {
        title: title || 'Untitled',
        content,
        category,
        tags: JSON.stringify(tags),
        is_draft: isDraft,
        is_pinned: isPinned,
        is_private: isPrivate,
        summary,
        slug: slug || undefined,
      }

      if (id) {
        await api.posts.update(id, payload as Record<string, unknown>)
      } else if (!post) {
        const created = await api.posts.create(payload)
        setPost(created)
        navigate(`/admin/posts/${created.id}/edit`, { replace: true })
      }
    } catch (err) {
      console.error('Auto-save failed:', err)
    } finally {
      setSaving(false)
    }
  }

  const handleSave = () => {
    if (saveTimer.current) clearTimeout(saveTimer.current)
    autoSave(getContent())
  }

  const handlePublish = async () => {
    handleSave()
    if (!post?.id) return
    await api.posts.publish(post.id, true)
    setIsDraft(false)
  }

  const handleUnpublish = async () => {
    if (!post?.id) return
    await api.posts.publish(post.id, false)
    setIsDraft(true)
  }

  const insertMarkdown = (before: string, after: string = '') => {
    const view = viewRef.current
    if (!view) return
    const selection = view.state.selection.main
    const selectedText = view.state.doc.sliceString(selection.from, selection.to)
    view.dispatch({
      changes: {
        from: selection.from,
        to: selection.to,
        insert: before + selectedText + after,
      },
      selection: { anchor: selection.from + before.length, head: selection.from + before.length + selectedText.length },
    })
    view.focus()
  }

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    try {
      const m = await api.media.upload(file)
      insertMarkdown(`![${file.name}](/${m.file_path})`)
    } catch (err) {
      console.error('Upload failed:', err)
    }
  }

  const toolbarButtons = [
    { label: 'B', title: '加粗', action: () => insertMarkdown('**', '**') },
    { label: 'I', title: '斜体', action: () => insertMarkdown('*', '*') },
    { label: '~~', title: '删除线', action: () => insertMarkdown('~~', '~~') },
    { label: 'H2', title: '标题', action: () => insertMarkdown('## ') },
    { label: '>', title: '引用', action: () => insertMarkdown('> ') },
    { label: '`', title: '行内代码', action: () => insertMarkdown('`', '`') },
    { label: '```', title: '代码块', action: () => insertMarkdown('```\n', '\n```') },
    { label: '•', title: '无序列表', action: () => insertMarkdown('- ') },
    { label: '1.', title: '有序列表', action: () => insertMarkdown('1. ') },
    { label: '🔗', title: '链接', action: () => insertMarkdown('[', '](url)') },
    { label: '—', title: '分割线', action: () => insertMarkdown('\n---\n') },
  ]

  return (
    <div className="h-[calc(100vh-4rem)] flex flex-col">
      {/* Top bar */}
      <div className="flex items-center justify-between mb-4 flex-shrink-0">
        <div className="flex items-center gap-3">
          <button onClick={() => navigate('/admin/posts')} className="text-[var(--color-text-secondary)] hover:text-[var(--color-text)]">
            ← 返回
          </button>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="文章标题..."
            className="text-xl font-semibold bg-transparent border-none text-[var(--color-text)] placeholder-[#484f58] focus:outline-none w-96"
          />
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-[var(--color-text-secondary)]">
            {saving ? '保存中...' : '已保存'}
          </span>
          <button
            onClick={() => setShowPreview(!showPreview)}
            className={`px-3 py-1.5 rounded-md text-xs border border-[var(--color-border)] transition-colors duration-150 ${
              showPreview ? 'bg-[#21262d] text-[var(--color-text)]' : 'text-[var(--color-text-secondary)]'
            }`}
          >
            {showPreview ? '隐藏预览' : '显示预览'}
          </button>
          {isDraft ? (
            <button
              onClick={handlePublish}
              className="px-3 py-1.5 rounded-md text-xs font-medium text-white bg-[#3fb950] hover:opacity-90 transition-all duration-150"
            >
              发布
            </button>
          ) : (
            <button
              onClick={handleUnpublish}
              className="px-3 py-1.5 rounded-md text-xs bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text-secondary)]"
            >
              取消发布
            </button>
          )}
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-1 mb-2 flex-shrink-0 flex-wrap">
        {toolbarButtons.map((btn) => (
          <button
            key={btn.title}
            title={btn.title}
            onClick={btn.action}
            className="w-7 h-7 flex items-center justify-center rounded text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:bg-[#21262d] transition-colors duration-150"
          >
            {btn.label}
          </button>
        ))}
        <label className="w-7 h-7 flex items-center justify-center rounded text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:bg-[#21262d] cursor-pointer transition-colors duration-150" title="上传图片">
          🖼
          <input type="file" accept="image/*" onChange={handleImageUpload} className="hidden" />
        </label>
      </div>

      {/* Editor + Preview + Sidebar */}
      <div className="flex-1 flex gap-4 min-h-0">
        {/* Editor */}
        <div className={`${showPreview ? 'flex-1' : 'flex-1'} min-w-0 border border-[var(--color-border)] rounded-lg overflow-hidden bg-[#0d1117]`}>
          <div ref={editorRef} className="h-full" />
        </div>

        {/* Preview */}
        {showPreview && (
          <div className="flex-1 min-w-0 border border-[var(--color-border)] rounded-lg overflow-auto bg-[var(--color-surface)]">
            <div
              className="p-6 prose prose-invert max-w-none text-sm leading-relaxed"
              style={{
                '--tw-prose-body': 'var(--color-text)',
                '--tw-prose-headings': 'var(--color-text)',
                '--tw-prose-links': 'var(--color-accent)',
                '--tw-prose-code': 'var(--color-text)',
                '--tw-prose-pre-bg': '#0d1117',
              } as React.CSSProperties}
              dangerouslySetInnerHTML={{ __html: preview || '<p class="text-[#484f58]">预览将在此显示...</p>' }}
            />
          </div>
        )}

        {/* Sidebar */}
        <div className="w-60 flex-shrink-0 space-y-4">
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-4 space-y-4">
            <div>
              <label className="block text-xs text-[var(--color-text-secondary)] mb-1">分类</label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="w-full px-2 py-1.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm focus:border-[var(--color-accent)]"
              >
                {categories.map(({ value, label }) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs text-[var(--color-text-secondary)] mb-1">标签</label>
              <input
                type="text"
                value={tagsInput}
                onChange={(e) => setTagsInput(e.target.value)}
                placeholder="tag1, tag2"
                className="w-full px-2 py-1.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)]"
              />
            </div>

            <div>
              <label className="block text-xs text-[var(--color-text-secondary)] mb-1">Slug</label>
              <input
                type="text"
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                placeholder="自动生成"
                className="w-full px-2 py-1.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)]"
              />
            </div>

            <div>
              <label className="block text-xs text-[var(--color-text-secondary)] mb-1">摘要</label>
              <textarea
                value={summary}
                onChange={(e) => setSummary(e.target.value)}
                placeholder="自动提取前200字"
                rows={3}
                className="w-full px-2 py-1.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm placeholder-[#484f58] focus:border-[var(--color-accent)] resize-none"
              />
            </div>

            <div className="space-y-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={isPinned} onChange={(e) => setIsPinned(e.target.checked)} className="accent-[var(--color-accent)]" />
                <span className="text-sm text-[var(--color-text-secondary)]">置顶</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={isPrivate} onChange={(e) => setIsPrivate(e.target.checked)} className="accent-[var(--color-accent)]" />
                <span className="text-sm text-[var(--color-text-secondary)]">私密</span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
