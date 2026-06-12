import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import { useAuth } from '../../contexts/AuthContext'
import PaginationBar from '../../components/common/PaginationBar'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const Wallet = () => {
  const { hasPermission } = useAuth()
  const [wallet, setWallet] = useState(null)
  const [transactions, setTransactions] = useState([])
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 10 })
  const [loading, setLoading] = useState(false)
  const [typeInput, setTypeInput] = useState('')
  const [type, setType] = useState('')
  const [directionInput, setDirectionInput] = useState('')
  const [direction, setDirection] = useState('')
  const canViewAllWallets = hasPermission('wallets', 'list')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [walletRes, txRes] = await Promise.all([
        walletService.getMyWallet(),
        walletService.getMyTransactions({
          page,
          limit: 10,
          'filters[type]': type || undefined,
          'filters[direction]': direction || undefined,
        }),
      ])
      const payload = getListPayload(txRes)
      setWallet(walletRes.data.data)
      setTransactions(payload.rows)
      setPagination(payload)
    } finally {
      setLoading(false)
    }
  }, [direction, page, type])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load wallet')))
  }, [load])

  const submitSearch = (event) => {
    event.preventDefault()
    setType(typeInput)
    setDirection(directionInput)
    setPage(1)
  }

  const resetSearch = () => {
    setTypeInput('')
    setType('')
    setDirectionInput('')
    setDirection('')
    setPage(1)
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>My Wallet</h1>
          <p>View your balance and complete transaction history.</p>
        </div>
        <div className="toolbar-actions">
          {canViewAllWallets && (
            <Link className="btn btn-outline-dark" to="/admin/wallets">
              <i className="bi bi-safe-fill me-2"></i>Admin Review
            </Link>
          )}
          <Link className="btn btn-primary d-flex align-items-center gap-2" to="/deposits">
             <i className="bi bi-plus-circle-fill"></i> Add Balance
          </Link>
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
             <i className="bi bi-wallet2"></i> Available Balance
          </div>
          <h2 className="luxe-hero-title">{formatMoney(wallet?.balance)}</h2>
          <p className="luxe-hero-subtitle">Currency: <strong>{wallet?.currency || 'IDR'}</strong></p>
        </div>
        <div className="luxe-hero-badge">
           <span className={`status-badge ${wallet?.is_active !== false ? 'success' : 'danger'}`}>
              {wallet?.is_active !== false ? 'Verified & Active' : 'Account Locked'}
           </span>
        </div>
      </div>

      <div className="toolbar-actions list-filter-bar">
        <form className="filter-pill filter-only compact-filter wallet-filter" onSubmit={submitSearch}>
          <i className="bi bi-funnel"></i>
          <select value={typeInput} onChange={(event) => setTypeInput(event.target.value)}>
            <option value="">All Transaction Types</option>
            <option value="deposit">Deposit (Topup)</option>
            <option value="debit_order">Service Order</option>
            <option value="refund_order">Refund Credit</option>
            <option value="adjustment">Manual Adjustment</option>
          </select>
          <select value={directionInput} onChange={(event) => setDirectionInput(event.target.value)}>
            <option value="">All Directions</option>
            <option value="credit">Credit (+)</option>
            <option value="debit">Debit (-)</option>
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
              <th>Type & Reference</th>
              <th>Flow</th>
              <th>Amount</th>
              <th>Transaction Date</th>
            </tr>
          </thead>
          <tbody>
            {transactions.map((row) => (
              <tr key={row.id}>
                <td>
                   <span className="table-main">
                      <strong className="text-capitalize">{row.type?.replace('_', ' ')}</strong>
                      <span className="table-subtext d-block"><i className="bi bi-hash"></i> {row.reference || row.deposit_request_id || '-'}</span>
                   </span>
                </td>
                <td>
                   <span className={`status-badge status-badge-sm ${row.direction === 'credit' ? 'success' : 'danger'} text-capitalize`}>
                      <i className={`bi bi-${row.direction === 'credit' ? 'arrow-down-left' : 'arrow-up-right'} me-1`}></i> {row.direction}
                   </span>
                </td>
                <td className="table-number">
                   <span className={`luxe-value-strong ${row.direction === 'credit' ? 'text-success' : 'text-danger'}`}>
                      {row.direction === 'credit' ? '+' : '-'}{formatMoney(row.amount)}
                   </span>
                </td>
                <td className="table-date"><i className="bi bi-clock me-1 text-muted"></i> {row.created_at ? new Date(row.created_at).toLocaleString('id-ID') : '-'}</td>
              </tr>
            ))}
            {!transactions.length && (
              <tr>
                <td colSpan="4" className="empty-cell py-5">
                   <div className="text-center">
                      <i className="bi bi-journal-x text-muted display-4"></i>
                      <p className="mt-3 text-muted">{loading ? 'Scanning ledger...' : 'No transactions recorded yet.'}</p>
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

export default Wallet
