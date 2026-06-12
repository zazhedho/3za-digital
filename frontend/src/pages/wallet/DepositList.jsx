import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import DepositForm from './DepositForm'
import { useAuth } from '../../contexts/AuthContext'
import { depositPayableAmount, depositProviderLabel, depositStatus, depositStatusClass, formatMoney } from '../../utils/deposit'
import PaginationBar from '../../components/common/PaginationBar'
import TableActionMenu from '../../components/common/TableActionMenu'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual Deposit'
  if (method === 'payment_gateway') return 'Payment Gateway'
  return method || 'Deposit Request'
}

const DepositList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 10 })
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
        limit: 10,
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
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>My Deposits</h1>
          <p>Create and track your wallet top-up requests.</p>
        </div>
        <div className="toolbar-actions">
          {canViewAllDeposits && (
            <Link className="btn btn-outline-dark" to="/admin/deposits">
              <i className="bi bi-bank me-2"></i>Admin Review
            </Link>
          )}
          {canCreateDeposit && (
            <button className="btn btn-primary d-flex align-items-center gap-2" type="button" onClick={() => setShowCreate(true)}>
              <i className="bi bi-plus-circle-fill"></i> New Deposit
            </button>
          )}
        </div>
      </div>

      {showCreate && (
        <DepositForm onClose={() => setShowCreate(false)} onCreated={handleCreated} />
      )}

      <div className="toolbar-actions list-filter-bar">
        <form className="filter-pill filter-only compact-filter status-filter" onSubmit={submitSearch}>
          <i className="bi bi-funnel"></i>
          <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
            <option value="">All Payment Status</option>
            <option value="pending">Pending</option>
            <option value="paid">Paid / Settled</option>
            <option value="expired">Expired</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <button className="btn btn-dark" type="submit" disabled={loading}>Filter</button>
          <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
            <i className="bi bi-x-lg me-2"></i>Reset
          </button>
        </form>
      </div>

      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Reference & Method</th>
              <th>Amount</th>
              <th>Final Payable</th>
              <th>Provider</th>
              <th>Status</th>
              <th>Date</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <span className="table-main">
                    <strong>{methodLabel(row.method)}</strong>
                    <span className="table-subtext d-block"><i className="bi bi-hash"></i> {row.payment_reference || 'Ref pending...'}</span>
                  </span>
                </td>
                <td className="table-number"><strong>{formatMoney(row.amount)}</strong></td>
                <td className="table-number">
                   <span className="luxe-value-strong text-primary">
                      {formatMoney(depositPayableAmount(row))}
                   </span>
                </td>
                <td className="table-nowrap"><span className="status-badge status-badge-sm info text-capitalize">{depositProviderLabel(row)}</span></td>
                <td>
                   <span className={`status-badge ${depositStatusClass(row)} text-capitalize`}>
                      {depositStatus(row)}
                   </span>
                </td>
                <td className="table-date"><i className="bi bi-clock me-1 text-muted"></i> {row.created_at ? new Date(row.created_at).toLocaleString('id-ID') : '-'}</td>
                <td className="text-end">
                  <TableActionMenu
                    label="Deposit actions"
                    items={[
                      { label: 'View Details', to: `/deposits/${row.id}` },
                    ]}
                  />
                </td>
              </tr>
            ))}
            {!rows.length && (
              <tr>
                <td colSpan="7" className="empty-cell py-5">
                   <div className="text-center">
                      <i className="bi bi-wallet2 text-muted display-4"></i>
                      <p className="mt-3 text-muted">{loading ? 'Scanning history...' : 'No deposits found.'}</p>
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

export default DepositList
