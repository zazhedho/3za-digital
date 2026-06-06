import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import DepositForm from './DepositForm'
import { useAuth } from '../../contexts/AuthContext'
import { depositPayableAmount, depositProviderLabel, depositStatus, depositStatusClass, formatMoney, isQRISDeposit } from '../../utils/deposit'
import PaginationBar from '../../components/common/PaginationBar'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual deposit'
  if (method === 'payment_gateway') return 'Payment gateway'
  return method || 'Deposit request'
}

const DepositList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 30 })
  const [loading, setLoading] = useState(false)
  const [statusInput, setStatusInput] = useState('')
  const [status, setStatus] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const canCreateDeposit = hasPermission('deposits', 'create')
  const canViewAllDeposits = hasPermission('admin_deposits', 'list')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await walletService.getMyDeposits({
        page,
        limit: 30,
        'filters[status]': status || undefined,
      })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } finally {
      setLoading(false)
    }
  }, [page, status])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposits')))
  }, [load])

  const handleCreated = async () => {
    if (page !== 1) {
      setPage(1)
      return
    }
    await load()
  }

  const submitSearch = (event) => {
    event.preventDefault()
    setStatus(statusInput)
    setPage(1)
  }

  const resetSearch = () => {
    setStatusInput('')
    setStatus('')
    setPage(1)
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Deposits</h1>
          <p>Create deposit requests and track status.</p>
        </div>
        <div className="toolbar-actions">
          {canViewAllDeposits && (
            <Link className="btn btn-outline-dark" to="/admin/deposits">
              <i className="bi bi-bank me-2"></i>Admin Deposits
            </Link>
          )}
          {canCreateDeposit && (
            <button className="btn btn-primary" type="button" onClick={() => setShowCreate(true)}>
              <i className="bi bi-plus-lg me-2"></i>Create deposit
            </button>
          )}
        </div>
      </div>

      {showCreate && (
        <DepositForm onClose={() => setShowCreate(false)} onCreated={handleCreated} />
      )}

      <form className="filter-pill filter-only status-filter" onSubmit={submitSearch}>
        <i className="bi bi-funnel"></i>
        <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
          <option value="">All status</option>
          <option value="pending">Pending</option>
          <option value="paid">Paid</option>
          <option value="expired">Expired</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <button className="btn btn-dark" type="submit" disabled={loading}>Filter</button>
        <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
          <i className="bi bi-x-lg me-2"></i>Reset
        </button>
      </form>

      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Reference</th>
              <th>Amount</th>
              <th>Pay</th>
              <th>Provider</th>
              <th>Status</th>
              <th>Date</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <span className="table-main">
                    <strong>{methodLabel(row.method)}</strong>
                    <span className="table-subtext">{row.payment_reference ? `Ref ${row.payment_reference}` : 'Waiting for reference'}</span>
                  </span>
                </td>
                <td className="table-number">{formatMoney(row.amount)}</td>
                <td className="table-number">{isQRISDeposit(row) ? formatMoney(depositPayableAmount(row)) : '-'}</td>
                <td className="table-nowrap">{depositProviderLabel(row)}</td>
                <td><span className={`status-badge ${depositStatusClass(row)} text-capitalize`}>{depositStatus(row)}</span></td>
                <td className="table-date">{row.created_at ? new Date(row.created_at).toLocaleString('id-ID') : '-'}</td>
                <td className="text-end">
                  <span className="table-actions">
                    <Link className="btn btn-sm btn-outline-dark" to={`/deposits/${row.id}`}>Detail</Link>
                  </span>
                </td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="7" className="empty-cell">{loading ? 'Loading...' : 'No deposits found'}</td></tr>}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default DepositList
