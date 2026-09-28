import { NavLink, Outlet } from 'react-router-dom'

const navItems = [
  { to: '/dashboard', label: '总览看板' },
  { to: '/dashboard/trend', label: '趋势分析' },
  { to: '/dashboard/intersections', label: '路口排行' },
  { to: '/alerts', label: '告警' },
  { to: '/imports', label: '数据导入' },
  { to: '/operations', label: '任务与告警管理' },
]

export default function Layout() {
  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="brand">交通数据分析平台</div>
        <nav>
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => (isActive ? 'nav-item active' : 'nav-item')}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}
