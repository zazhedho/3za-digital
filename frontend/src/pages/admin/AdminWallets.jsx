import { useCallback, useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import { useAuth } from '../../contexts/AuthContext'
import { formatMoney } from '../../utils/deposit'

const AdminWallets = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [pagination, setPagination] = useState(null)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [activeWallet, setActiveWallet] = useState(null)
  const [confirmAdjust, setConfirmAdjust] = useState(false)
  const [confirmLoading, setConfirmLoading] = useState(false)
  const [form, setForm] = useState({ amount: '', direction: 'credit', description: '' })
  const canAdjust = hasPermission('wallets', 'adjust')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await walletService.getWallets({ page, limit: 50 })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } finally {
      setLoading(false)
    }
  }, [page])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load wallets')))
  }, [load])

  const adjust = async (event) => {
    event.preventDefault()
    setConfirmAdjust(true)
  }

  const openAdjust = (wallet) => {
    setActiveWallet(wallet)
    setForm({ amount: '', direction: 'credit', description: '' })
    setConfirmAdjust(false)
  }

  const closeAdjust = () => {
    setActiveWallet(null)
    setConfirmAdjust(false)
    setForm({ amount: '', direction: 'credit', description: '' })
  }

  const confirmAdjustment = async () => {
    setConfirmLoading(true)
    try {
      await walletService.adminAdjust(activeWallet.user_id, form)
      toast.success('Wallet adjusted')
      closeAdjust()
      setConfirmAdjust(false)
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Adjustment failed'))
    } finally {
      setConfirmLoading(false)
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Admin Wallets</h1>
          <p>Wallet balances and manual adjustments.</p>
        </div>
      </div>

      {activeWallet && (
        <div className="modal-backdrop-custom" role="dialog" aria-modal="true">
          <div className="modal-panel wallet-adjust-modal">
            <div className="modal-heading">
              <div>
                <h5>Adjust Wallet</h5>
                <p>{activeWallet.user?.name || activeWallet.user?.email || activeWallet.user_id}</p>
              </div>
              <button className="modal-close" type="button" onClick={closeAdjust} aria-label="Close">
                <i className="bi bi-x-lg"></i>
              </button>
            </div>
            <form onSubmit={adjust}>
              <div className="payment-summary mb-3">
                <div><span>Current balance</span><strong>{formatMoney(activeWallet.balance)}</strong></div>
                <div><span>Currency</span><strong>{activeWallet.currency || 'IDR'}</strong></div>
              </div>

              <label className="form-label">Direction</label>
              <div className="segmented-control mb-3">
                <button
                  type="button"
                  className={form.direction === 'credit' ? 'active' : ''}
                  onClick={() => setForm({ ...form, direction: 'credit' })}
                >
                  Credit
                </button>
                <button
                  type="button"
                  className={form.direction === 'debit' ? 'active danger' : ''}
                  onClick={() => setForm({ ...form, direction: 'debit' })}
                >
                  Debit
                </button>
              </div>

              <label className="form-label">Amount</label>
              <input
                className="form-control"
                type="number"
                min="1"
                placeholder="Enter amount"
                value={form.amount}
                onChange={(event) => setForm({ ...form, amount: event.target.value })}
                required
              />

              <label className="form-label mt-3">Description</label>
              <textarea
                className="form-control"
                rows="3"
                placeholder="Adjustment reason"
                value={form.description}
                onChange={(event) => setForm({ ...form, description: event.target.value })}
                required
              />

              <div className="toolbar-actions justify-content-end mt-4">
                <button className="btn btn-outline-dark" type="button" onClick={closeAdjust}>Cancel</button>
                <button className={form.direction === 'debit' ? 'btn btn-danger' : 'btn btn-primary'} type="submit">Review Adjustment</button>
              </div>
            </form>
          </div>
        </div>
      )}

      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>User</th>
              <th>Balance</th>
              <th>Currency</th>
              <th>Status</th>
              {canAdjust && <th className="text-end">Action</th>}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <span className="table-main">
                    <strong>{row.user?.name || row.user_id}</strong>
                    {row.user?.email && <span className="table-subtext">{row.user.email}</span>}
                  </span>
                </td>
                <td className="table-number">{formatMoney(row.balance)}</td>
                <td className="table-nowrap">{row.currency}</td>
                <td><span className={`badge ${row.is_active ? 'text-bg-success' : 'text-bg-secondary'}`}>{row.is_active ? 'Active' : 'Inactive'}</span></td>
                {canAdjust && (
                  <td className="text-end">
                    <span className="table-actions">
                      <button className="btn btn-sm btn-outline-dark" onClick={() => openAdjust(row)}>Adjust</button>
                    </span>
                  </td>
                )}
              </tr>
            ))}
            {!rows.length && <tr><td colSpan={canAdjust ? 5 : 4} className="empty-cell">{loading ? 'Loading...' : 'No wallets found'}</td></tr>}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
      <ConfirmationModal
        show={confirmAdjust}
        title="Adjust Wallet"
        message={`${form.direction === 'debit' ? 'Debit' : 'Credit'} ${formatMoney(form.amount)} for ${activeWallet?.user?.name || activeWallet?.user?.email || 'this user'}?`}
        confirmLabel="Save Adjustment"
        confirmClassName={form.direction === 'debit' ? 'btn-danger' : 'btn-primary'}
        loading={confirmLoading}
        onCancel={() => setConfirmAdjust(false)}
        onConfirm={confirmAdjustment}
      />
    </div>
  )
}

export default AdminWallets
