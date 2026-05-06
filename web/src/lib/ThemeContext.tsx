import { createContext, useContext, useEffect, useState, useCallback } from 'react'
import { api } from './api'

interface ThemeColors {
  bg: string
  surface: string
  text: string
  textSecondary: string
  accent: string
  border: string
  codeBg: string
}

interface ThemeFonts {
  body: string
  code: string
}

interface ThemeState {
  name: string
  colors: ThemeColors
  fonts: ThemeFonts
  customCSS: string
  footerText: string
}

interface ThemeContextValue {
  theme: ThemeState
  setPreset: (name: 'dark' | 'light' | 'ink') => void
  updateColors: (colors: Partial<ThemeColors>) => void
  updateFonts: (fonts: Partial<ThemeFonts>) => void
  setCustomCSS: (css: string) => void
  presets: Record<string, { colors: ThemeColors; fonts: ThemeFonts }>
}

const PRESETS: Record<string, { colors: ThemeColors; fonts: ThemeFonts }> = {
  dark: {
    colors: { bg: '#0d1117', surface: '#161b22', text: '#c9d1d9', textSecondary: '#8b949e', accent: '#58a6ff', border: '#30363d', codeBg: '#0d1117' },
    fonts: { body: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif', code: '"JetBrains Mono", "Fira Code", monospace' },
  },
  light: {
    colors: { bg: '#ffffff', surface: '#f6f8fa', text: '#1f2328', textSecondary: '#656d76', accent: '#0969da', border: '#d0d7de', codeBg: '#f6f8fa' },
    fonts: { body: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif', code: '"JetBrains Mono", "Fira Code", monospace' },
  },
  ink: {
    colors: { bg: '#f5f2eb', surface: '#ede8db', text: '#2c3e2d', textSecondary: '#6b7c6d', accent: '#d97706', border: '#d4cbb8', codeBg: '#ede8db' },
    fonts: { body: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif', code: '"JetBrains Mono", "Fira Code", monospace' },
  },
}

function applyColors(colors: ThemeColors) {
  const root = document.documentElement
  root.style.setProperty('--color-bg', colors.bg)
  root.style.setProperty('--color-surface', colors.surface)
  root.style.setProperty('--color-text', colors.text)
  root.style.setProperty('--color-text-secondary', colors.textSecondary)
  root.style.setProperty('--color-accent', colors.accent)
  root.style.setProperty('--color-border', colors.border)
  root.style.setProperty('--color-code-bg', colors.codeBg)
}

function applyFonts(fonts: ThemeFonts) {
  const root = document.documentElement
  root.style.setProperty('--font-body', fonts.body)
  root.style.setProperty('--font-code', fonts.code)
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<ThemeState>({
    name: 'dark',
    colors: PRESETS.dark.colors,
    fonts: PRESETS.dark.fonts,
    customCSS: '',
    footerText: '',
  })

  // Inject custom CSS
  useEffect(() => {
    let styleEl = document.getElementById('mo-custom-css') as HTMLStyleElement | null
    if (!styleEl) {
      styleEl = document.createElement('style')
      styleEl.id = 'mo-custom-css'
      document.head.appendChild(styleEl)
    }
    styleEl.textContent = theme.customCSS || ''
  }, [theme.customCSS])

  // Load saved theme from localStorage, then fetch from server
  useEffect(() => {
    const saved = localStorage.getItem('mo_theme')
    if (saved) {
      try {
        const parsed = JSON.parse(saved)
        applyColors(parsed.colors)
        applyFonts(parsed.fonts)
        setTheme(parsed)
      } catch {}
    }

    // Fetch public settings from server to sync
    api.settings.public().then((res: any) => {
      const next: ThemeState = {
        name: res.theme?.name || 'dark',
        colors: { ...PRESETS.dark.colors },
        fonts: { ...PRESETS.dark.fonts },
        customCSS: res.custom_css || '',
        footerText: res.footer_text || '',
      }

      // Apply theme colors from server or defaults
      if (res.theme?.accent_color) {
        const preset = PRESETS[res.theme.name] || PRESETS.dark
        next.colors = { ...preset.colors }
        if (res.theme.accent_color) next.colors.accent = res.theme.accent_color
      }
      if (res.theme?.font_body) next.fonts.body = res.theme.font_body === 'system'
        ? PRESETS.dark.fonts.body
        : res.theme.font_body
      if (res.theme?.font_code) next.fonts.code = res.theme.font_code === 'jetbrains-mono'
        ? PRESETS.dark.fonts.code
        : res.theme.font_code

      // Only override localStorage if no saved theme
      if (!saved) {
        applyColors(next.colors)
        applyFonts(next.fonts)
        setTheme(next)
      }
    }).catch(() => {})
  }, [])

  const setPreset = useCallback((name: 'dark' | 'light' | 'ink') => {
    const preset = PRESETS[name]
    const next: ThemeState = {
      ...theme,
      name,
      colors: { ...preset.colors },
      fonts: { ...preset.fonts },
    }
    applyColors(next.colors)
    applyFonts(next.fonts)
    setTheme(next)
    localStorage.setItem('mo_theme', JSON.stringify(next))
  }, [theme])

  const updateColors = useCallback((colors: Partial<ThemeColors>) => {
    setTheme(prev => {
      const next = { ...prev, colors: { ...prev.colors, ...colors }, name: 'custom' }
      applyColors(next.colors)
      localStorage.setItem('mo_theme', JSON.stringify(next))
      return next
    })
  }, [])

  const updateFonts = useCallback((fonts: Partial<ThemeFonts>) => {
    setTheme(prev => {
      const next = { ...prev, fonts: { ...prev.fonts, ...fonts }, name: 'custom' }
      applyFonts(next.fonts)
      localStorage.setItem('mo_theme', JSON.stringify(next))
      return next
    })
  }, [])

  const setCustomCSS = useCallback((css: string) => {
    setTheme(prev => ({ ...prev, customCSS: css }))
  }, [])

  return (
    <ThemeContext.Provider value={{ theme, setPreset, updateColors, updateFonts, setCustomCSS, presets: PRESETS }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}
