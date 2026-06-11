import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import auditService from '../../services/auditService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'
import TableActionMenu from '../../components/common/TableActionMenu'

const formatDate = (value) => {
  if (!value) return '-'
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

const statusClass = (status) => {
  if (status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  return 'warning'
}

const AuditList = () => {
  const [params, setParams] = useSearchParams()
  const [rows, setRows] = useState([])
  const [searchInput, setSearchInput] = useState(params.get('search') || '')
  const [search, setSearch] = useState(params.get('search') || '')
  const [statusInput, setStatusInput] = useState('')
  const [status, setStatus] = useState('')
  const [actionInput, setActionInput] = useState('')
  const [action, setAction] = useState('')
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 20 })
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const nextSearch = params.get('search') || ''
    setSearchInput(nextSearch)
    setSearch(nextSearch)
    setPage(1)
  }, [params])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await auditService.getAll({
        search,
        page,
        limit: 20,
        'filters[status]': status || undefined,
        'filters[action]': action || undefined,
      })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } finally {
      setLoading(false)
    }
  }, [action, page, search, status])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load audit trails')))
  }, [load])

  const submitSearch = (event) => {
    event.preventDefault()
    const nextSearch = searchInput.trim()
    setSearch(nextSearch)
    setStatus(statusInput)
    setAction(actionInput)
    setPage(1)
    setParams(nextSearch ? { search: nextSearch } : {})
  }

  const resetSearch = () => {
    setSearchInput('')
    setSearch('')
    setStatusInput('')
    setStatus('')
    setActionInput('')
    setAction('')
    setPage(1)
    setParams({})
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Audit Trails</h1>
          <p>System activity tracking and security forensic logs.</p>
        </div>
      </div>

      <div className="toolbar-actions mb-4">
        <form className="filter-pill filter-only" onSubmit={submitSearch}>
          <i className="bi bi-funnel"></i>
          <input
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            placeholder="Search message or request ID..."
            style={{ minWidth: '240px' }}
          />
          <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
            <option value="">All Status</option>
            <option value="success">Success</option>
            <option value="failed">Failed</option>
            <option value="pending">Pending</option>
          </select>
          <select value={actionInput} onChange={(event) => setActionInput(event.target.value)}>
            <option value="">All Actions</option>
            <option value="create">Create</option>
            <option value="update">Update</option>
            <option value="delete">Delete</option>
            <option value="login">Login</option>
            <option value="logout">Logout</option>
          </select>
          <button className="btn btn-dark" type="submit" disabled={loading}>Search</button>
          <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
            <i className="bi bi-x-lg me-2"></i>Reset
          </button>
        </form>
      </div>

      <section className="table-panel">
        <table className="table app-table table-wide align-middle">
          <thead>
            <tr>
              <th>Event Summary</th>
              <th>System Actor</th>
              <th>Resource Context</th>
              <th>Status</th>
              <th>Occurred At</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <span className="table-main">
                    <strong className="d-block text-truncate" style={{ maxWidth: '280px' }}>{row.summary || row.message || 'Audit event'}</strong>
                    <code className="table-subtext" style={{ fontSize: '11px' }}>ID: {row.request_id || row.id.split('-')[0]}</code>
                  </span>
                </td>
                <td>
                  <span className="table-main">
                    <strong>{row.actor?.role || 'System'}</strong>
                    <span className="table-subtext d-block">{row.actor?.user_id || 'Internal'}</span>
                  </span>
                </td>
                <td>
                   <div className="table-main">
                      <span className="table-subtext d-block"><strong>{row.resource_label || row.resource || '-'}</strong></span>
                      <small className="status-badge status-badge-sm info text-capitalize">{row.action_label || row.action || '-'}</small>
                   </div>
                </td>
                <td><span className={`status-badge ${statusClass(row.status)} text-capitalize`}>{row.status_label || row.status || '-'}</span></td>
                <td className="table-date"><i className="bi bi-clock me-1 text-muted"></i> {formatDate(row.occurred_at)}</td>
                <td className="text-end">
                   <TableActionMenu
                      label="Audit actions"
                      items={[
                        { label: 'View Forensic Detail', to: `/audits/${row.id}` },
                        { label: 'View Actor Profile', to: row.actor?.user_id ? `/users/${row.actor.user_id}` : null, hidden: !row.actor?.user_id },
                      ]}
                   />
                </td>
              </tr>
            ))}
            {!rows.length && (
               <tr>
                  <td colSpan="6" className="empty-cell py-5">
                    <div className="text-center">
                       <i className="bi bi-shield-shaded text-muted display-4"></i>
                       <p className="mt-3 text-muted">{loading ? 'Scanning logs...' : 'No audit records found.'}</p>
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

export default AuditList
