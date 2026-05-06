import { useState, useEffect } from 'react'
import { useTheme } from '../lib/ThemeContext'
import { settings } from '../lib/api'

const TABS = [
  { key: 'site', label: '站点' },
  { key: 'appearance', label: '外观' },
  { key: 'advanced', label: '高级' },
]

export default function SettingsPage() {
  const [tab, setTab] = useState('site')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  // Site form
  const [siteTitle, setSiteTitle] = useState('')
  const [siteSubtitle, setSiteSubtitle] = useState('')
  const [siteDesc, setSiteDesc] = useState('')

  // Appearance form
  const { theme, setPreset, updateColors, updateFonts, presets } = useTheme()
  const [accentColor, setAccentColor] = useState(theme.colors.accent)
  const [bgColor, setBgColor] = useState(theme.colors.bg)
  const [surfaceColor, setSurfaceColor] = useState(theme.colors.surface)
  const [textColor, setTextColor] = useState(theme.colors.text)
  const [textSecondaryColor, setTextSecondaryColor] = useState(theme.colors.textSecondary)
  const [borderColor, setBorderColor] = useState(theme.colors.border)
  const [codeBgColor, setCodeBgColor] = useState(theme.colors.codeBg)
  const [fontBody, setFontBody] = useState('system')
  const [fontCode, setFontCode] = useState('jetbrains-mono')

  // Advanced
  const [customCSS, setCustomCSS] = useState('')
  const [customJS, setCustomJS] = useState('')
  const [footerText, setFooterText] = useState('')
  const [navItems, setNavItems] = useState('')

  useEffect(() => {
    settings.get().then((res: any) => {
      const s = res.settings
      setSiteTitle(s.site?.title || '')
      setSiteSubtitle(s.site?.subtitle || '')
      setSiteDesc(s.site?.description || '')
      if (s.theme?.accent_color) setAccentColor(s.theme.accent_color)
      if (s.theme?.font_body) setFontBody(s.theme.font_body)
      if (s.theme?.font_code) setFontCode(s.theme.font_code)
      setCustomCSS(s.custom_css || '')
      setCustomJS(s.custom_js || '')
      setFooterText(s.footer_text || '')
      setNavItems(s.nav_items || '')
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [])

  const handlePreset = (name: string) => {
    setPreset(name as 'dark' | 'light' | 'ink')
    const p = presets[name]
    if (p) {
      setAccentColor(p.colors.accent)
      setBgColor(p.colors.bg)
      setSurfaceColor(p.colors.surface)
      setTextColor(p.colors.text)
      setTextSecondaryColor(p.colors.textSecondary)
      setBorderColor(p.colors.border)
      setCodeBgColor(p.colors.codeBg)
    }
  }

  const handleApplyColors = () => {
    updateColors({ bg: bgColor, surface: surfaceColor, text: textColor, textSecondary: textSecondaryColor, accent: accentColor, border: borderColor, codeBg: codeBgColor })
  }

  const handleApplyFonts = () => {
    const bodyFont = fontBody === 'system' ? presets.dark.fonts.body : fontBody === 'serif' ? 'Georgia, "Times New Roman", serif' : presets.dark.fonts.body
    const codeFont = fontCode === 'jetbrains-mono' ? presets.dark.fonts.code : fontCode === 'fira-code' ? '"Fira Code", monospace' : '"Source Code Pro", monospace'
    updateFonts({ body: bodyFont, code: codeFont })
  }

  const handleSaveSite = async () => {
    setSaving(true)
    setMessage('')
    try {
      await settings.update({
        'site.title': siteTitle,
        'site.subtitle': siteSubtitle,
        'site.description': siteDesc,
      })
      setMessage('站点设置已保存')
    } catch (err: any) { setMessage(err.message || '保存失败') }
    setSaving(false)
  }

  const handleSaveAppearance = async () => {
    setSaving(true)
    setMessage('')
    try {
      await settings.update({
        'theme.name': theme.name,
        'theme.accent_color': accentColor,
        'theme.font_body': fontBody,
        'theme.font_code': fontCode,
      })
      setMessage('外观设置已保存')
    } catch (err: any) { setMessage(err.message || '保存失败') }
    setSaving(false)
  }

  const handleSaveAdvanced = async () => {
    setSaving(true)
    setMessage('')
    try {
      await settings.update({
        'custom.css': customCSS,
        'custom.js': customJS,
        'footer.text': footerText,
        'nav.items': navItems,
      })
      setMessage('高级设置已保存')
    } catch (err: any) { setMessage(err.message || '保存失败') }
    setSaving(false)
  }

  if (loading) {
    return <div className="text-center text-[var(--color-text-secondary)] py-12">加载中...</div>
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold text-[var(--color-text)] mb-6">设置</h1>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-[var(--color-border)]">
        {TABS.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`px-4 py-2 text-sm rounded-t-md transition-colors duration-150 ${
              tab === key
                ? 'bg-[var(--color-surface)] text-[var(--color-text)] border border-b-0 border-[var(--color-border)]'
                : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Message */}
      {message && (
        <div className={`mb-4 px-4 py-2 rounded-md text-sm ${
          message.includes('失败') ? 'bg-red-900/20 text-red-300 border border-red-800' : 'bg-green-900/20 text-green-300 border border-green-800'
        }`}>
          {message}
        </div>
      )}

      {/* Tab: Site */}
      {tab === 'site' && (
        <div className="space-y-6">
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">站点信息</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-xs text-[var(--color-text-secondary)] mb-1.5">站点标题</label>
                <input type="text" value={siteTitle} onChange={e => setSiteTitle(e.target.value)}
                  className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm focus:border-[var(--color-accent)]" />
              </div>
              <div>
                <label className="block text-xs text-[var(--color-text-secondary)] mb-1.5">副标题</label>
                <input type="text" value={siteSubtitle} onChange={e => setSiteSubtitle(e.target.value)}
                  className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm focus:border-[var(--color-accent)]" />
              </div>
              <div>
                <label className="block text-xs text-[var(--color-text-secondary)] mb-1.5">站点描述 (SEO)</label>
                <textarea value={siteDesc} onChange={e => setSiteDesc(e.target.value)} rows={2}
                  className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm focus:border-[var(--color-accent)] resize-none" />
              </div>
            </div>
          </div>
          <button onClick={handleSaveSite} disabled={saving}
            className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150">
            {saving ? '保存中...' : '保存站点设置'}
          </button>
        </div>
      )}

      {/* Tab: Appearance */}
      {tab === 'appearance' && (
        <div className="space-y-6">
          {/* Theme presets */}
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">主题预设</h2>
            <div className="grid grid-cols-3 gap-3">
              {(['dark', 'light', 'ink'] as const).map(name => {
                const p = presets[name]
                return (
                  <button
                    key={name}
                    onClick={() => handlePreset(name)}
                    className={`p-4 rounded-lg border-2 transition-all duration-150 text-left ${
                      theme.name === name ? 'border-[var(--color-accent)]' : 'border-[var(--color-border)] hover:border-[#484f58]'
                    }`}
                    style={{ background: p.colors.bg }}
                  >
                    <div className="flex gap-1.5 mb-2">
                      <div className="w-3 h-3 rounded-full" style={{ background: p.colors.accent }} />
                      <div className="w-3 h-3 rounded-full" style={{ background: p.colors.textSecondary }} />
                      <div className="w-3 h-3 rounded-full" style={{ background: p.colors.border }} />
                    </div>
                    <div className="text-xs font-medium" style={{ color: p.colors.text }}>
                      {name === 'dark' ? '暗夜' : name === 'light' ? '素白' : '墨绿'}
                    </div>
                    <div className="text-xs mt-0.5" style={{ color: p.colors.textSecondary }}>
                      {name.charAt(0).toUpperCase() + name.slice(1)}
                    </div>
                  </button>
                )
              })}
            </div>
          </div>

          {/* Colors */}
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">自定义配色</h2>
            <div className="grid grid-cols-3 gap-4">
              {[
                { label: '强调色', value: accentColor, set: setAccentColor },
                { label: '背景色', value: bgColor, set: setBgColor },
                { label: '表面色', value: surfaceColor, set: setSurfaceColor },
                { label: '文字色', value: textColor, set: setTextColor },
                { label: '次要文字', value: textSecondaryColor, set: setTextSecondaryColor },
                { label: '边框色', value: borderColor, set: setBorderColor },
                { label: '代码块背景', value: codeBgColor, set: setCodeBgColor },
              ].map(({ label, value, set }) => (
                <div key={label}>
                  <label className="block text-xs text-[var(--color-text-secondary)] mb-1.5">{label}</label>
                  <div className="flex items-center gap-2">
                    <input type="color" value={value} onChange={e => set(e.target.value)}
                      className="w-8 h-8 rounded-md border border-[var(--color-border)] cursor-pointer bg-transparent p-0" />
                    <input type="text" value={value} onChange={e => set(e.target.value)}
                      className="flex-1 px-2 py-1.5 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-xs font-mono focus:border-[var(--color-accent)]" />
                  </div>
                </div>
              ))}
            </div>
            <button onClick={handleApplyColors}
              className="mt-4 px-3 py-1.5 rounded-md text-xs font-medium bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text)] hover:text-white hover:bg-[#30363d] transition-colors duration-150">
              应用配色
            </button>
          </div>

          {/* Fonts */}
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">字体设置</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs text-[var(--color-text-secondary)] mb-1.5">正文字体</label>
                <select value={fontBody} onChange={e => setFontBody(e.target.value)}
                  className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm focus:border-[var(--color-accent)]">
                  <option value="system">系统默认</option>
                  <option value="serif">衬线 (Serif)</option>
                  <option value="sans-serif">无衬线 (Sans-serif)</option>
                </select>
              </div>
              <div>
                <label className="block text-xs text-[var(--color-text-secondary)] mb-1.5">代码字体</label>
                <select value={fontCode} onChange={e => setFontCode(e.target.value)}
                  className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm focus:border-[var(--color-accent)]">
                  <option value="jetbrains-mono">JetBrains Mono</option>
                  <option value="fira-code">Fira Code</option>
                  <option value="system-mono">系统等宽</option>
                </select>
              </div>
            </div>
            <button onClick={handleApplyFonts}
              className="mt-4 px-3 py-1.5 rounded-md text-xs font-medium bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text)] hover:text-white hover:bg-[#30363d] transition-colors duration-150">
              应用字体
            </button>
          </div>

          <button onClick={handleSaveAppearance} disabled={saving}
            className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150">
            {saving ? '保存中...' : '保存外观设置'}
          </button>
        </div>
      )}

      {/* Tab: Advanced */}
      {tab === 'advanced' && (
        <div className="space-y-6">
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">自定义 CSS</h2>
            <p className="text-xs text-[var(--color-text-secondary)] mb-2">插入到前台 &lt;head&gt; 中，可用于统计代码或个性化调整。</p>
            <textarea value={customCSS} onChange={e => setCustomCSS(e.target.value)} rows={8}
              className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm font-mono focus:border-[var(--color-accent)] resize-none" />
          </div>

          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">自定义 JS</h2>
            <p className="text-xs text-[var(--color-text-secondary)] mb-2">插入到前台 &lt;/body&gt; 之前。</p>
            <textarea value={customJS} onChange={e => setCustomJS(e.target.value)} rows={6}
              className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm font-mono focus:border-[var(--color-accent)] resize-none" />
          </div>

          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">页脚文字</h2>
            <p className="text-xs text-[var(--color-text-secondary)] mb-2">支持 HTML，如备案号。</p>
            <input type="text" value={footerText} onChange={e => setFooterText(e.target.value)}
              className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm focus:border-[var(--color-accent)]" />
          </div>

          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
            <h2 className="text-sm font-medium text-[var(--color-text)] mb-4">导航菜单</h2>
            <p className="text-xs text-[var(--color-text-secondary)] mb-2">JSON 数组格式: [{"{"}"label": "首页", "url": "/"{"}"}]</p>
            <textarea value={navItems} onChange={e => setNavItems(e.target.value)} rows={4}
              className="w-full px-3 py-2 rounded-md bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text)] text-sm font-mono focus:border-[var(--color-accent)] resize-none" />
          </div>

          <button onClick={handleSaveAdvanced} disabled={saving}
            className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150">
            {saving ? '保存中...' : '保存高级设置'}
          </button>
        </div>
      )}
    </div>
  )
}
