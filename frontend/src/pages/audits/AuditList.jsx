import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import auditService from '../../services/auditService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'

const formatDate = (value) => {
  if (!value) return '-'
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

const statusClass = (status) => {
  if (status === 'success') return 'status-paid'
  if (status === 'failed') return 'status-failed'
  if (status === 'pending') return 'status-pending'
  return 'status-default'
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
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Audit Trails</h1>
          <p>System activity history and request outcomes.</p>
        </div>
      </div>

      <form className="filter-pill" onSubmit={submitSearch}>
        <i className="bi bi-search"></i>
        <input
          value={searchInput}
          onChange={(event) => setSearchInput(event.target.value)}
          placeholder="Search action, resource, request ID, or message"
        />
        <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
          <option value="">All status</option>
          <option value="success">Success</option>
          <option value="failed">Failed</option>
          <option value="pending">Pending</option>
        </select>
        <select value={actionInput} onChange={(event) => setActionInput(event.target.value)}>
          <option value="">All actions</option>
          <option value="create">Create</option>
          <option value="update">Update</option>
          <option value="delete">Delete</option>
          <option value="login">Login</option>
          <option value="logout">Logout</option>
          <option value="refresh">Refresh</option>
          <option value="assign">Assign</option>
        </select>
        <button className="btn btn-dark" type="submit" disabled={loading}>Search</button>
        <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
          <i className="bi bi-x-lg me-2"></i>Reset
        </button>
      </form>

      <section className="table-panel">
        <table className="table app-table table-wide align-middle">
          <thead>
            <tr>
              <th>Summary</th>
              <th>Actor</th>
              <th>Resource</th>
              <th>Status</th>
              <th>Occurred</th>
              <th className="text-end">Action</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <span className="table-main">
                    <strong>{row.summary || row.message || 'Audit event'}</strong>
                    <span className="table-subtext">{row.request_id ? `Request ${row.request_id}` : row.ip_address || '-'}</span>
                  </span>
                </td>
                <td>
                  <span className="table-main">
                    <strong>{row.actor?.role || 'User'}</strong>
                    <span className="table-subtext">{row.actor?.user_id || 'System'}</span>
                  </span>
                </td>
                <td>
                  <span className="table-main">
                    <strong>{row.resource_label || row.resource || '-'}</strong>
                    <span className="table-subtext">{row.action_label || row.action || '-'}</span>
                  </span>
                </td>
                <td><span className={`status-badge ${statusClass(row.status)}`}>{row.status_label || row.status || '-'}</span></td>
                <td className="table-date">{formatDate(row.occurred_at)}</td>
                <td className="text-end">
                  <Link className="btn btn-outline-dark btn-sm" to={`/audits/${row.id}`}>
                    Detail
                  </Link>
                </td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="6" className="empty-cell">{loading ? 'Loading...' : 'No audit trails found'}</td></tr>}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default AuditList
