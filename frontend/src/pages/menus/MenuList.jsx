import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import menuService from '../../services/menuService'
import { getErrorMessage, getListPayload } from '../../services/api'
import TableActionMenu from '../../components/common/TableActionMenu'
import PaginationBar from '../../components/common/PaginationBar'

const MenuList = () => {
  const { hasPermission } = useAuth()
  const [params, setParams] = useSearchParams()
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [searchInput, setSearchInput] = useState(params.get('search') || '')
  const [search, setSearch] = useState(params.get('search') || '')
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 10 })

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await menuService.getAll({ search, page, limit: 10 })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } finally {
      setLoading(false)
    }
  }, [search, page])

  useEffect(() => {
    const nextSearch = params.get('search') || ''
    setSearchInput(nextSearch)
    setSearch(nextSearch)
    setPage(1)
  }, [params])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load menus')))
  }, [load])

  const submitSearch = (event) => {
    event.preventDefault()
    const nextSearch = searchInput.trim()
    setSearch(nextSearch)
    setPage(1)
    setParams(nextSearch ? { search: nextSearch } : {})
  }

  const resetSearch = () => {
    setSearchInput('')
    setSearch('')
    setPage(1)
    setParams({})
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Navigation Menus</h1>
          <p>Configure sidebar navigation and application route visibility.</p>
        </div>
      </div>

      <div className="toolbar-actions mb-4">
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
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default MenuList
