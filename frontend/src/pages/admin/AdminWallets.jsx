import { useCallback, useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import { useAuth } from '../../contexts/AuthContext'
import { formatMoney } from '../../utils/deposit'
import TableActionMenu from '../../components/common/TableActionMenu'

const AdminWallets = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [pagination, setPagination] = useState(null)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [currencyInput, setCurrencyInput] = useState('')
  const [currency, setCurrency] = useState('')
  const [statusInput, setStatusInput] = useState('')
  const [isActive, setIsActive] = useState('')
  const [activeWallet, setActiveWallet] = useState(null)
  const [confirmAdjust, setConfirmAdjust] = useState(false)
  const [confirmLoading, setConfirmLoading] = useState(false)
  const [form, setForm] = useState({ amount: '', direction: 'credit', description: '' })
  const canAdjust = hasPermission('wallets', 'adjust')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await walletService.getWallets({
        page,
        limit: 50,
        'filters[currency]': currency || undefined,
        'filters[is_active]': isActive || undefined,
      })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } finally {
      setLoading(false)
    }
  }, [currency, isActive, page])

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
      toast.success('Wallet adjusted successfully')
      closeAdjust()
      setConfirmAdjust(false)
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Adjustment failed'))
    } finally {
      setConfirmLoading(false)
    }
  }

  const submitSearch = (event) => {
    event.preventDefault()
    setCurrency(currencyInput)
    setIsActive(statusInput)
    setPage(1)
  }

  const resetSearch = () => {
    setCurrencyInput('')
    setCurrency('')
    setStatusInput('')
    setIsActive('')
    setPage(1)
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Admin Wallets</h1>
          <p>Monitor user balances and perform manual balance adjustments.</p>
        </div>
      </div>

      <div className="toolbar-actions list-filter-bar">
        <form className="filter-pill filter-only compact-filter status-filter" onSubmit={submitSearch}>
          <i className="bi bi-funnel"></i>
          <select value={currencyInput} onChange={(event) => setCurrencyInput(event.target.value)}>
            <option value="">All Currencies</option>
            <option value="IDR">IDR</option>
          </select>
          <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
            <option value="">All Status</option>
            <option value="true">Active Only</option>
            <option value="false">Inactive Only</option>
          </select>
          <button className="btn btn-dark" type="submit" disabled={loading}>Filter</button>
          <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
            <i className="bi bi-x-lg me-2"></i>Reset
          </button>
        </form>
      </div>

      {activeWallet && (
        <div className="modal-backdrop-custom" role="dialog" aria-modal="true">
          <div className="modal-panel">
            <div className="modal-heading">
              <div>
                <h5>Adjust Balance</h5>
                <p>User: <strong>{activeWallet.user?.name || activeWallet.user?.email}</strong></p>
              </div>
              <button className="modal-close" type="button" onClick={closeAdjust} aria-label="Close">
                <i className="bi bi-x-lg"></i>
              </button>
            </div>
            <form onSubmit={adjust} className="deposit-form-modern">
              <div className="deposit-fee-card mb-3">
                <div className="deposit-fee-row">
                   <span>Current Balance</span>
                   <strong>{formatMoney(activeWallet.balance)}</strong>
                </div>
                <div className="deposit-fee-row">
                   <span>Currency</span>
                   <span className="status-badge status-badge-sm info">{activeWallet.currency || 'IDR'}</span>
                </div>
              </div>

              <div className="deposit-input-group">
                 <label>Direction</label>
                 <div className="segmented-control">
                    <button
                      type="button"
                      className={form.direction === 'credit' ? 'active' : ''}
                      onClick={() => setForm({ ...form, direction: 'credit' })}
                    >
                      <i className="bi bi-plus-circle me-2"></i> Credit (Add)
                    </button>
                    <button
                      type="button"
                      className={form.direction === 'debit' ? 'active danger' : ''}
                      onClick={() => setForm({ ...form, direction: 'debit' })}
                    >
                      <i className="bi bi-dash-circle me-2"></i> Debit (Subtract)
                    </button>
                 </div>
              </div>

              <div className="deposit-input-group mt-3">
                 <label>Amount</label>
                 <div className="deposit-amount-wrapper">
                    <span className="deposit-amount-currency">Rp</span>
                    <input
                      type="number"
                      min="1"
                      placeholder="0"
                      value={form.amount}
                      onChange={(event) => setForm({ ...form, amount: event.target.value })}
                      required
                    />
                 </div>
              </div>

              <div className="deposit-input-group mt-3">
                 <label>Adjustment Note</label>
                 <textarea
                    className="form-control"
                    rows="2"
                    placeholder="Reason for this manual adjustment..."
                    value={form.description}
                    onChange={(event) => setForm({ ...form, description: event.target.value })}
                    required
                    style={{ borderRadius: '14px', padding: '12px' }}
                 />
              </div>

              <div className="deposit-modal-actions mt-4 pt-3 border-top">
                <button className="btn btn-outline-dark" type="button" onClick={closeAdjust}>Cancel</button>
                <button className={form.direction === 'debit' ? 'btn btn-danger' : 'btn btn-primary'} type="submit">
                   Review & Save
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Member Account</th>
              <th>Current Balance</th>
              <th>Currency</th>
              <th>Status</th>
              {canAdjust && <th className="text-end">Actions</th>}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <div className="d-flex align-items-center gap-3">
                    <div className="user-avatar-luxe" style={{ width: '36px', height: '36px', borderRadius: '10px' }}>
                       <div className="avatar-placeholder" style={{ fontSize: '14px' }}>{row.user?.name?.charAt(0) || '?'}</div>
                    </div>
                    <span className="table-main">
                      <strong>{row.user?.name || 'Anonymous User'}</strong>
                      <span className="table-subtext d-block">{row.user?.email || row.user_id}</span>
                    </span>
                  </div>
                </td>
                <td className="table-number">
                   <span className="luxe-value-strong">{formatMoney(row.balance)}</span>
                </td>
                <td><span className="status-badge status-badge-sm info">{row.currency}</span></td>
                <td>
                   <span className={`status-badge ${row.is_active ? 'success' : 'danger'}`}>
                      {row.is_active ? 'Active' : 'Inactive'}
                   </span>
                </td>
                {canAdjust && (
                  <td className="text-end">
                     <TableActionMenu
                        label="Wallet actions"
                        items={[
                          { label: 'Adjust Balance', onClick: () => openAdjust(row) },
                          { label: 'View Transactions', to: `/audits?filters[resource]=wallet&filters[resource_id]=${row.id}` },
                        ]}
                     />
                  </td>
                )}
              </tr>
            ))}
            {!rows.length && (
               <tr>
                  <td colSpan={canAdjust ? 5 : 4} className="empty-cell py-5">
                    <div className="text-center">
                       <i className="bi bi-wallet2 text-muted display-4"></i>
                       <p className="mt-3 text-muted">{loading ? 'Loading wallets...' : 'No wallets found.'}</p>
                    </div>
                  </td>
               </tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
      
      <ConfirmationModal
        show={confirmAdjust}
        title="Confirm Adjustment"
        message={`You are about to ${form.direction} ${formatMoney(form.amount)} to ${activeWallet?.user?.name}'s wallet. This action will be recorded in the audit trail.`}
        confirmLabel="Confirm & Save"
        confirmClassName={form.direction === 'debit' ? 'btn-danger' : 'btn-primary'}
        loading={confirmLoading}
        onCancel={() => setConfirmAdjust(false)}
        onConfirm={confirmAdjustment}
      />
    </div>
  )
}

export default AdminWallets
