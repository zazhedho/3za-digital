import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { useAuth } from '../../contexts/AuthContext'

const titleMap = {
  dashboard: 'Dashboard',
  smm: 'SMM',
  services: 'Services',
  orders: 'Orders',
  new: 'Create',
  wallet: 'Wallet',
  deposits: 'Deposits',
  admin: 'Admin',
  users: 'Users',
  roles: 'Roles',
  configs: 'Configs',
  profile: 'Profile',
  edit: 'Edit',
}

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

const getRouteTitle = (parts) => {
  const path = `/${parts.join('/')}`
  if (uuidPattern.test(parts.at(-1) || '')) {
    if (path.startsWith('/admin/deposits/')) return 'Admin Deposit Detail'
    if (path.startsWith('/smm/orders/')) return 'SMM Order Detail'
    if (path.startsWith('/deposits/')) return 'Deposit Detail'
    if (path.startsWith('/users/')) return 'User Detail'
    if (path.startsWith('/roles/')) return 'Role Detail'
    if (path.startsWith('/configs/')) return 'Config Detail'
    if (path.startsWith('/menus/')) return 'Menu Detail'
    return 'Detail'
  }
  if (parts.at(-1) === 'edit') {
    const resource = titleMap[parts.at(-3)] || titleMap[parts.at(-2)] || 'Record'
    return `Edit ${resource.replace(/s$/, '')}`
  }
  return parts.length ? titleMap[parts.at(-1)] || parts.at(-1) : 'Dashboard'
}

const getCrumbLabel = (crumb) => (uuidPattern.test(crumb) ? 'Detail' : (titleMap[crumb] || crumb))

const TopNav = ({ onToggleMobileMenu, isMobileMenuOpen }) => {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [search, setSearch] = useState('')

  const parts = location.pathname.split('/').filter(Boolean)
  const title = getRouteTitle(parts)

  useEffect(() => {
    const query = new URLSearchParams(location.search)
    setSearch(query.get('search') || '')
  }, [location.search])

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const crumbs = ['dashboard', ...parts.filter((part) => part !== 'dashboard')]

  const submitSearch = (event) => {
    event.preventDefault()
    const value = search.trim()
    if (!value) return

    const targetPath = location.pathname.startsWith('/smm/orders') ? '/smm/orders' : '/smm/services'
    navigate(`${targetPath}?search=${encodeURIComponent(value)}`)
  }

  return (
    <header className="modern-topnav">
      <button
        type="button"
        className="mobile-menu-toggle"
        onClick={onToggleMobileMenu}
        aria-label={isMobileMenuOpen ? 'Close sidebar menu' : 'Open sidebar menu'}
      >
        <i className={`bi ${isMobileMenuOpen ? 'bi-x-lg' : 'bi-list'}`}></i>
      </button>

      <div className="topnav-left">
        <h4 className="page-title">{title}</h4>
        <nav aria-label="breadcrumb">
          <ol className="breadcrumb mb-0">
            {crumbs.map((crumb, index) => (
              <li className={`breadcrumb-item ${index === crumbs.length - 1 ? 'active' : ''}`} key={`${crumb}-${index}`}>
                {index === crumbs.length - 1 ? (
                  <span>{getCrumbLabel(crumb)}</span>
                ) : (
                  <Link to="/dashboard">{getCrumbLabel(crumb)}</Link>
                )}
              </li>
            ))}
          </ol>
        </nav>
      </div>

      <form className="topnav-search" onSubmit={submitSearch}>
        <i className="bi bi-search"></i>
        <input
          type="search"
          placeholder="Search services or orders"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
        />
        <button className="topnav-search-button" type="submit" aria-label="Search">
          <i className="bi bi-arrow-right"></i>
        </button>
      </form>

      <div className="topnav-user">
        <div className="user-info">
          <span className="user-name">{user?.name || user?.email || 'User'}</span>
          <span className="user-role">{user?.role || 'member'}</span>
        </div>
        <button className="avatar-button" type="button" data-bs-toggle="dropdown" aria-expanded="false">
          <i className="bi bi-person"></i>
        </button>
        <ul className="dropdown-menu dropdown-menu-end">
          <li><Link className="dropdown-item" to="/profile">Profile</Link></li>
          <li><button className="dropdown-item text-danger" type="button" onClick={handleLogout}>Logout</button></li>
        </ul>
      </div>
    </header>
  )
}

export default TopNav
