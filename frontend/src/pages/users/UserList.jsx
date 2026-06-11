import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import userService from '../../services/userService'
import { getErrorMessage, getListPayload } from '../../services/api'
import TableActionMenu from '../../components/common/TableActionMenu'

const UserList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await userService.getAll({ search, limit: 50 })
      setRows(getListPayload(response).rows)
    } finally {
      setLoading(false)
    }
  }, [search])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load users')))
  }, [load])

  const submitSearch = (event) => {
    event.preventDefault()
    setSearch(searchInput.trim())
  }

  const resetSearch = () => {
    setSearchInput('')
    setSearch('')
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Users</h1>
          <p>Manage registered accounts and system roles.</p>
        </div>
        {hasPermission('users', 'create') && (
          <Link to="/users/new" className="btn btn-primary d-flex align-items-center gap-2">
            <i className="bi bi-person-plus-fill"></i> New User
          </Link>
        )}
      </div>

      <div className="toolbar-actions list-filter-bar">
        <form className="filter-pill filter-only" onSubmit={submitSearch}>
          <i className="bi bi-search"></i>
          <input 
             value={searchInput} 
             onChange={(event) => setSearchInput(event.target.value)} 
             placeholder="Search by name or email..." 
          />
          <button className="btn btn-dark" type="submit" disabled={loading}>Search</button>
          <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
            <i className="bi bi-x-lg me-2"></i>Reset
          </button>
        </form>
      </div>

      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Member</th>
              <th>Contact Info</th>
              <th>System Role</th>
              <th>Account Status</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <div className="d-flex align-items-center gap-3">
                    <div className="user-avatar-luxe" style={{ width: '38px', height: '38px', borderRadius: '10px' }}>
                       <div className="avatar-placeholder" style={{ fontSize: '14px' }}>{row.name?.charAt(0) || '?'}</div>
                    </div>
                    <span className="table-main"><strong>{row.name || 'Anonymous'}</strong></span>
                  </div>
                </td>
                <td>
                   <div className="table-main">
                      <span className="table-subtext d-block"><i className="bi bi-envelope me-1"></i> {row.email}</span>
                      {row.phone && <span className="table-subtext d-block"><i className="bi bi-phone me-1"></i> {row.phone}</span>}
                   </div>
                </td>
                <td><span className="status-badge info text-capitalize">{row.role}</span></td>
                <td>
                   <span className={`status-badge ${row.is_active !== false ? 'success' : 'danger'}`}>
                      {row.is_active !== false ? 'Active' : 'Inactive'}
                   </span>
                </td>
                <td className="text-end">
                  <TableActionMenu
                    label="User actions"
                    items={[
                      { label: 'View Profile', to: `/users/${row.id}`, hidden: !hasPermission('users', 'view') },
                      { label: 'Edit Account', to: `/users/${row.id}/edit`, hidden: !hasPermission('users', 'update') },
                      { label: 'Login History', to: `/audits?filters[resource]=user&filters[resource_id]=${row.id}`, hidden: !hasPermission('audits', 'list') },
                    ]}
                  />
                </td>
              </tr>
            ))}
            {!rows.length && (
              <tr>
                <td colSpan="5" className="empty-cell py-5">
                   <div className="text-center">
                      <i className="bi bi-people text-muted display-4"></i>
                      <p className="mt-3 text-muted">{loading ? 'Searching users...' : 'No users found.'}</p>
                   </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default UserList
