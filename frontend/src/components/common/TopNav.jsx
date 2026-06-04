import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useState } from 'react'
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
}

const TopNav = ({ onToggleMobileMenu, isMobileMenuOpen }) => {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [search, setSearch] = useState('')

  const parts = location.pathname.split('/').filter(Boolean)
  const title = parts.length ? titleMap[parts.at(-1)] || parts.at(-1) : 'Dashboard'

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const crumbs = ['dashboard', ...parts.filter((part) => part !== 'dashboard')]

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
                  <span>{titleMap[crumb] || crumb}</span>
                ) : (
                  <Link to="/dashboard">{titleMap[crumb] || crumb}</Link>
                )}
              </li>
            ))}
          </ol>
        </nav>
      </div>

      <div className="topnav-search">
        <i className="bi bi-search"></i>
        <input
          type="search"
          placeholder="Search services or orders"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter' && search.trim()) {
              navigate(`/smm/services?search=${encodeURIComponent(search.trim())}`)
            }
          }}
        />
      </div>

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
