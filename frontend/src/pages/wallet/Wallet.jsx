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
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 30 })
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
          limit: 30,
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
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Wallet</h1>
          <p>Balance and transaction ledger.</p>
        </div>
        {canViewAllWallets && (
          <Link className="btn btn-outline-dark" to="/admin/wallets">
            <i className="bi bi-safe me-2"></i>Admin Wallets
          </Link>
        )}
      </div>

      <div className="wallet-card">
        <span>Available balance</span>
        <strong>{formatMoney(wallet?.balance)}</strong>
        <small>{wallet?.currency || 'IDR'} - {wallet?.is_active === false ? 'Inactive' : 'Active'}</small>
      </div>

      <form className="filter-pill filter-only wallet-filter" onSubmit={submitSearch}>
        <i className="bi bi-funnel"></i>
        <select value={typeInput} onChange={(event) => setTypeInput(event.target.value)}>
          <option value="">All types</option>
          <option value="deposit">Deposit</option>
          <option value="debit_order">Order</option>
          <option value="refund_order">Refund order</option>
          <option value="adjustment">Adjustment</option>
        </select>
        <select value={directionInput} onChange={(event) => setDirectionInput(event.target.value)}>
          <option value="">All directions</option>
          <option value="credit">Credit</option>
          <option value="debit">Debit</option>
        </select>
        <button className="btn btn-dark" type="submit" disabled={loading}>Filter</button>
        <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
          <i className="bi bi-x-lg me-2"></i>Reset
        </button>
      </form>

      <section className="table-panel mt-4">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Type</th>
              <th>Direction</th>
              <th>Amount</th>
              <th>Reference</th>
              <th>Date</th>
            </tr>
          </thead>
          <tbody>
            {transactions.map((row) => (
              <tr key={row.id}>
                <td className="text-capitalize">{row.type}</td>
                <td><span className="badge text-bg-light text-capitalize">{row.direction}</span></td>
                <td className="table-number">{formatMoney(row.amount)}</td>
                <td><span className="table-subtext">{row.reference || row.deposit_request_id || '-'}</span></td>
                <td className="table-date">{row.created_at ? new Date(row.created_at).toLocaleString('id-ID') : '-'}</td>
              </tr>
            ))}
            {!transactions.length && <tr><td colSpan="5" className="empty-cell">{loading ? 'Loading...' : 'No transactions found'}</td></tr>}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default Wallet
