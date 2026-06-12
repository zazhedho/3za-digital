import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import roleService from '../../services/roleService'
import { getErrorMessage, getListPayload } from '../../services/api'
import TableActionMenu from '../../components/common/TableActionMenu'

const RoleList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await roleService.getAll({ search, limit: 10 })
      setRows(getListPayload(response).rows)
    } finally {
      setLoading(false)
    }
  }, [search])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load roles')))
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
          <h1>Roles</h1>
          <p>Access control levels and permission registry.</p>
        </div>
        {hasPermission('roles', 'create') && (
          <Link to="/roles/new" className="btn btn-primary d-flex align-items-center gap-2">
            <i className="bi bi-shield-plus"></i> New Role
          </Link>
        )}
      </div>

      <div className="toolbar-actions list-filter-bar">
        <form className="filter-pill filter-only" onSubmit={submitSearch}>
          <i className="bi bi-search"></i>
          <input 
             value={searchInput} 
             onChange={(event) => setSearchInput(event.target.value)} 
             placeholder="Search by role name..." 
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
              <th>Role Name</th>
              <th>Display Name</th>
              <th>Description</th>
              <th>Origin</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                   <div className="d-flex align-items-center gap-3">
                      <div className="status-dot success" style={{ width: '8px', height: '8px' }}></div>
                      <span className="table-main"><code>{row.name}</code></span>
                   </div>
                </td>
                <td><span className="table-main"><strong>{row.display_name || '-'}</strong></span></td>
                <td><span className="table-subtext">{row.description || 'No description provided'}</span></td>
                <td>
                   <span className={`status-badge ${row.is_system ? 'info' : 'success'}`}>
                      {row.is_system ? 'System Protected' : 'User Custom'}
                   </span>
                </td>
                <td className="text-end">
                  <TableActionMenu
                    label="Role actions"
                    items={[
                      { label: 'View Details', to: `/roles/${row.id}`, hidden: !hasPermission('roles', 'view') },
                      { label: 'Edit Permissions', to: `/roles/${row.id}/edit`, hidden: !hasPermission('roles', 'update') },
                    ]}
                  />
                </td>
              </tr>
            ))}
            {!rows.length && (
               <tr>
                  <td colSpan="5" className="empty-cell py-5">
                    <div className="text-center">
                       <i className="bi bi-shield-lock text-muted display-4"></i>
                       <p className="mt-3 text-muted">{loading ? 'Loading roles...' : 'No roles defined yet.'}</p>
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

export default RoleList
