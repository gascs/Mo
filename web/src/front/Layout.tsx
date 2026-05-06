import { Outlet, NavLink } from 'react-router-dom'

const navItems = [
  { to: '/', label: '首页' },
  { to: '/?category=tech', label: '技术' },
  { to: '/?category=life', label: '生活' },
  { to: '/treehole', label: '树洞' },
  { to: '/archive', label: '归档' },
  { to: '/search', label: '搜索' },
  { to: '/about', label: '关于' },
  { to: '/links', label: '友链' },
]

export default function FrontLayout() {
  return (
    <div className="min-h-screen flex flex-col bg-[var(--color-bg)]">
      <header className="border-b border-[var(--color-border)] bg-[var(--color-surface)]">
        <div className="max-w-3xl mx-auto px-6 flex items-center justify-between h-14">
          <NavLink to="/" className="text-lg font-semibold text-[var(--color-accent)] tracking-[-0.02em]">
            Mo
          </NavLink>
          <nav className="flex items-center gap-1">
            {navItems.map(({ to, label }) => (
              <NavLink
                key={to}
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `px-3 py-1.5 text-sm rounded-md transition-colors duration-150 ${
                    isActive
                      ? 'text-[var(--color-text)]'
                      : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'
                  }`
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>

      <main className="flex-1">
        <Outlet />
      </main>

      <footer className="border-t border-[var(--color-border)] py-8 text-center">
        <p className="text-sm text-[var(--color-text-secondary)]">
          Mo Blog &copy; 2026 <a href="https://github.com/gascs" className="text-[var(--color-text-secondary)] hover:text-[var(--color-text)] transition-colors">@gascs</a>
        </p>
      </footer>
    </div>
  )
}
