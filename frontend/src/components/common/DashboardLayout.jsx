import { Link, Outlet, useLocation } from 'react-router-dom'
import { useEffect, useMemo, useState } from 'react'
import menuService from '../../services/menuService'
import { useAuth } from '../../contexts/AuthContext'
import TopNav from './TopNav'

const normalizeMenus = (menus) => (
  menus
    .filter((menu) => menu.path && menu.path !== '/profile')
    .map((menu) => ({
      path: menu.path,
      label: menu.display_name || menu.label || menu.name,
      icon: menu.icon || 'bi-circle',
      orderIndex: menu.order_index || 0,
    }))
    .sort((a, b) => a.orderIndex - b.orderIndex)
)

const DashboardLayout = () => {
  const location = useLocation()
  const { hasPermission } = useAuth()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [collapsed, setCollapsed] = useState(false)
  const [apiMenus, setApiMenus] = useState([])

  useEffect(() => {
    let mounted = true
    menuService.getUserMenus()
      .then((response) => {
        if (mounted) setApiMenus(normalizeMenus(response.data.data || []))
      })
      .catch(() => {
        if (mounted) setApiMenus([])
      })
    return () => {
      mounted = false
    }
  }, [])

  useEffect(() => {
    setMobileOpen(false)
  }, [location.pathname])

  const menus = useMemo(() => {
    const nextMenus = [...apiMenus]
    const addPermissionMenu = (resource, action, item) => {
      if (!hasPermission(resource, action) || nextMenus.some((menu) => menu.path === item.path)) return
      nextMenus.push(item)
    }

    addPermissionMenu('wallets', 'list', {
      path: '/admin/wallets',
      label: 'Admin Wallets',
      icon: 'bi-safe',
      orderIndex: 905,
    })
    addPermissionMenu('admin_deposits', 'list', {
      path: '/admin/deposits',
      label: 'Admin Deposits',
      icon: 'bi-bank',
      orderIndex: 906,
    })

    return nextMenus.sort((a, b) => a.orderIndex - b.orderIndex)
  }, [apiMenus, hasPermission])

  return (
    <div className={`layout-wrapper ${mobileOpen ? 'sidebar-open' : ''} ${collapsed ? 'sidebar-collapsed' : ''}`}>
      <button className="sidebar-toggle-btn" type="button" onClick={() => setCollapsed((value) => !value)}>
        <i className={`bi ${collapsed ? 'bi-chevron-right' : 'bi-chevron-left'}`}></i>
      </button>

      <div className="sidebar-overlay" onClick={() => setMobileOpen(false)}></div>

      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="brand-mark">3ZA</div>
          <div>
            <h3 className="sidebar-title">3ZA Digital</h3>
            <small className="sidebar-subtitle">Social commerce panel</small>
          </div>
        </div>

        <nav className="sidebar-menu">
          {menus.map((item) => {
            const active = location.pathname === item.path || location.pathname.startsWith(`${item.path}/`)
            return (
              <Link key={item.path} to={item.path} className={`sidebar-menu-item ${active ? 'active' : ''}`}>
                <i className={`bi ${item.icon}`}></i>
                <span className="menu-label">{item.label}</span>
              </Link>
            )
          })}
          {!menus.length && <div className="sidebar-empty">No permitted menus</div>}
        </nav>
      </aside>

      <main className="main-content">
        <TopNav onToggleMobileMenu={() => setMobileOpen((value) => !value)} isMobileMenuOpen={mobileOpen} />
        <div className="content-area">
          <Outlet />
        </div>
      </main>
    </div>
  )
}

export default DashboardLayout
