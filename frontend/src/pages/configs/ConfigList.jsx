import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import appConfigService from '../../services/appConfigService'
import { getErrorMessage, getListPayload } from '../../services/api'
import TableActionMenu from '../../components/common/TableActionMenu'
import PaginationBar from '../../components/common/PaginationBar'

const renderConfigValue = (value) => {
  const normalized = String(value ?? '').trim().toLowerCase()
  if (normalized === 'true') return <span className="status-badge status-badge-sm success">true</span>
  if (normalized === 'false') return <span className="status-badge status-badge-sm danger">false</span>
  if (!String(value ?? '').trim()) return <span className="table-subtext italic">Empty</span>
  return <code className="luxe-value-code">{String(value)}</code>
}

const ConfigList = () => {
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
      const response = await appConfigService.getAll({ search, page, limit: 10 })
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
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load configs')))
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
          <h1>App Configurations</h1>
          <p>Fine-tune system behavior and runtime parameters.</p>
        </div>
      </div>

      <div className="toolbar-actions list-filter-bar">
        <form className="filter-pill filter-only" onSubmit={submitSearch}>
          <i className="bi bi-search"></i>
          <input 
             value={searchInput} 
             onChange={(event) => setSearchInput(event.target.value)} 
             placeholder="Search by key or display name..." 
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
              <th>Config Name</th>
              <th>Config Key</th>
              <th>Category</th>
              <th>Current Value</th>
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
                      <span className="table-subtext d-block text-truncate" style={{ maxWidth: '240px' }}>{row.description || 'System setting'}</span>
                   </span>
                </td>
                <td><code className="luxe-value-code">{row.config_key}</code></td>
                <td><span className="status-badge status-badge-sm info text-capitalize">{row.category}</span></td>
                <td>{renderConfigValue(row.value)}</td>
                <td>
                   <span className={`status-badge ${row.is_active ? 'success' : 'danger'}`}>
                      {row.is_active ? 'Active' : 'Disabled'}
                   </span>
                </td>
                <td className="text-end">
                  <TableActionMenu
                    label="Config actions"
                    items={[
                      { label: 'View Detail', to: `/configs/${row.id}`, hidden: !hasPermission('configs', 'view') },
                      { label: 'Edit Value', to: `/configs/${row.id}/edit`, hidden: !hasPermission('configs', 'update') },
                    ]}
                  />
                </td>
              </tr>
            ))}
            {!rows.length && (
               <tr>
                  <td colSpan="6" className="empty-cell py-5">
                    <div className="text-center">
                       <i className="bi bi-gear-wide-connected text-muted display-4"></i>
                       <p className="mt-3 text-muted">{loading ? 'Loading settings...' : 'No configurations found.'}</p>
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

export default ConfigList
