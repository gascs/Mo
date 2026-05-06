/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: 'var(--color-bg)',
        surface: 'var(--color-surface)',
        text: 'var(--color-text)',
        'text-secondary': 'var(--color-text-secondary)',
        accent: 'var(--color-accent)',
        border: 'var(--color-border)',
        'code-bg': 'var(--color-code-bg)',
      },
      fontFamily: {
        body: ['var(--font-body)', 'system-ui', '-apple-system', 'sans-serif'],
        code: ['var(--font-code)', 'JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}
