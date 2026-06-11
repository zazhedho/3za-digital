import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import menuService from '../../services/menuService'
import { getErrorMessage, getListPayload } from '../../services/api'
import TableActionMenu from '../../components/common/TableActionMenu'

const MenuList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await menuService.getAll({ search, limit: 100 })
      setRows(getListPayload(response).rows)
    } finally {
      setLoading(false)
    }
  }, [search])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load menus')))
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
          <h1>Navigation Menus</h1>
          <p>Configure sidebar navigation and application route visibility.</p>
        </div>
        {hasPermission('menus', 'create') && (
          <Link to="/menus/new" className="btn btn-primary d-flex align-items-center gap-2">
            <i className="bi bi-layout-sidebar-inset"></i> New Menu
          </Link>
        )}
      </div>

      <div className="toolbar-actions list-filter-bar">
        <form className="filter-pill filter-only" onSubmit={submitSearch}>
          <i className="bi bi-search"></i>
          <input 
             value={searchInput} 
             onChange={(event) => setSearchInput(event.target.value)} 
             placeholder="Search by label or path..." 
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
              <th>Display Label</th>
              <th>Route Path</th>
              <th>Icon</th>
              <th>Sort Order</th>
              <th>Status</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                   <span className="table-main">
                      <strong>{row.display_name}</strong>
                      <code className="table-subtext d-block">{row.name}</code>
                   </span>
                </td>
                <td><span className="luxe-value-code">{row.path || '-'}</span></td>
                <td>
                   <div className="d-flex align-items-center gap-2">
                      <i className={`bi bi-${row.icon || 'circle'} text-primary`}></i>
                      <span className="table-subtext">{row.icon || '-'}</span>
                   </div>
                </td>
                <td className="table-number"><strong>{row.order_index}</strong></td>
                <td>
                   <span className={`status-badge ${row.is_active ? 'success' : 'danger'}`}>
                      {row.is_active ? 'Visible' : 'Hidden'}
                   </span>
                </td>
                <td className="text-end">
                  <TableActionMenu
                    label="Menu actions"
                    items={[
                      { label: 'View Detail', to: `/menus/${row.id}`, hidden: !hasPermission('menus', 'view') },
                      { label: 'Edit Menu', to: `/menus/${row.id}/edit`, hidden: !hasPermission('menus', 'update') },
                    ]}
                  />
                </td>
              </tr>
            ))}
            {!rows.length && (
               <tr>
                  <td colSpan="6" className="empty-cell py-5">
                    <div className="text-center">
                       <i className="bi bi-list-ul text-muted display-4"></i>
                       <p className="mt-3 text-muted">{loading ? 'Loading menus...' : 'No menu entries found.'}</p>
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

export default MenuList
