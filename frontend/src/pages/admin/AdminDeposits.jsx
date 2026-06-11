import { useCallback, useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import TableActionMenu from '../../components/common/TableActionMenu'
import { depositMetadata, depositPayableAmount, depositStatus, depositStatusClass, formatMoney } from '../../utils/deposit'
import { useAuth } from '../../contexts/AuthContext'
import PaginationBar from '../../components/common/PaginationBar'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual Review'
  if (method === 'payment_gateway') return 'Payment Gateway'
  return method || 'Deposit Request'
}

const AdminDeposits = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 50 })
  const [loading, setLoading] = useState(false)
  const [statusInput, setStatusInput] = useState('')
  const [status, setStatus] = useState('')
  const [confirmAction, setConfirmAction] = useState(null)
  const [confirmLoading, setConfirmLoading] = useState(false)
  const [cancelReason, setCancelReason] = useState('')
  const canApproveDeposit = hasPermission('admin_deposits', 'approve')
  const canCancelDeposit = hasPermission('admin_deposits', 'cancel')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await walletService.getDeposits({
        page,
        limit: 50,
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

  const approve = async () => {
    if (!confirmAction?.deposit) return
    setConfirmLoading(true)
    try {
      await walletService.approveDeposit(confirmAction.deposit.id, { amount: confirmAction.deposit.amount })
      toast.success('Deposit approved successfully')
      setConfirmAction(null)
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Approve failed'))
    } finally {
      setConfirmLoading(false)
    }
  }

  const cancel = async () => {
    if (!confirmAction?.deposit) return
    setConfirmLoading(true)
    try {
      await walletService.cancelDeposit(confirmAction.deposit.id, { reason: cancelReason.trim() })
      toast.success('Deposit rejected')
      setConfirmAction(null)
      setCancelReason('')
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Cancel failed'))
    } finally {
      setConfirmLoading(false)
    }
  }

  const confirmDeposit = confirmAction?.deposit
  const isCancelAction = confirmAction?.type === 'cancel'
  const openConfirm = (type, deposit) => {
    setCancelReason('')
    setConfirmAction({ type, deposit })
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
          <h1>Admin Deposits</h1>
          <p>Review and process user wallet top-up requests.</p>
        </div>
      </div>

      <div className="toolbar-actions mb-4">
        <form className="filter-pill filter-only status-filter" onSubmit={submitSearch}>
          <i className="bi bi-funnel"></i>
          <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
            <option value="">All Payment Status</option>
            <option value="pending">Pending Review</option>
            <option value="paid">Settled / Paid</option>
            <option value="expired">Expired</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Rejected / Cancelled</option>
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
              <th>Requesting User</th>
              <th>Reference & Method</th>
              <th className="text-end">Base Amount</th>
              <th className="text-end">Final Payable</th>
              <th>Status</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => {
              const metadata = depositMetadata(row)
              return (
                <tr key={row.id}>
                  <td>
                    <div className="d-flex align-items-center gap-3">
                      <div className="user-avatar-luxe" style={{ width: '36px', height: '36px', borderRadius: '10px' }}>
                         <div className="avatar-placeholder" style={{ fontSize: '14px' }}>{row.user?.name?.charAt(0) || '?'}</div>
                      </div>
                      <span className="table-main">
                        <strong>{row.user?.name || 'Unknown'}</strong>
                        <span className="table-subtext d-block">{row.user?.email}</span>
                      </span>
                    </div>
                  </td>
                  <td>
                    <span className="table-main">
                      <strong>{methodLabel(row.method)}</strong>
                      <span className="table-subtext d-block">
                         <i className="bi bi-hash"></i> {row.payment_reference || 'NO-REF'}
                         {metadata.cancel_reason && <span className="text-danger ms-2">• {metadata.cancel_reason}</span>}
                      </span>
                    </span>
                  </td>
                  <td className="table-number"><strong>{formatMoney(row.amount)}</strong></td>
                  <td className="table-number">
                     <span className="luxe-value-strong text-primary">
                        {formatMoney(depositPayableAmount(row))}
                     </span>
                  </td>
                  <td>
                     <span className={`status-badge ${depositStatusClass(row)} text-capitalize`}>
                        {depositStatus(row)}
                     </span>
                  </td>
                  <td className="text-end">
                    <TableActionMenu
                      label="Deposit actions"
                      items={[
                        { label: 'Review Details', to: `/admin/deposits/${row.id}` },
                        { label: 'Approve Payment', hidden: !canApproveDeposit || row.status !== 'pending', onClick: () => openConfirm('approve', row) },
                        { label: 'Reject Request', hidden: !canCancelDeposit || row.status !== 'pending', danger: true, onClick: () => openConfirm('cancel', row) },
                      ]}
                    />
                  </td>
                </tr>
              )
            })}
            {!rows.length && (
               <tr>
                  <td colSpan="6" className="empty-cell py-5">
                    <div className="text-center">
                       <i className="bi bi-wallet2 text-muted display-4"></i>
                       <p className="mt-3 text-muted">{loading ? 'Scanning deposits...' : 'No deposit requests found.'}</p>
                    </div>
                  </td>
               </tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
      
      <ConfirmationModal
        show={Boolean(confirmAction)}
        title={isCancelAction ? 'Reject Deposit Request' : 'Approve Payment'}
        message={isCancelAction 
           ? `Are you sure you want to reject the deposit for ${confirmDeposit?.user?.name}? This action cannot be undone.`
           : `Confirm receipt of ${formatMoney(depositPayableAmount(confirmDeposit))} from ${confirmDeposit?.user?.name}? Wallet will be credited ${formatMoney(confirmDeposit?.amount)}.`
        }
        confirmLabel={isCancelAction ? 'Reject Now' : 'Approve Now'}
        confirmClassName={isCancelAction ? 'btn-danger' : 'btn-primary'}
        loading={confirmLoading}
        confirmDisabled={isCancelAction && !cancelReason.trim()}
        onCancel={() => {
          setConfirmAction(null)
          setCancelReason('')
        }}
        onConfirm={isCancelAction ? cancel : approve}
      >
        {isCancelAction && (
          <div className="mt-3">
            <label className="form-label">Rejection Reason</label>
            <textarea
              className="form-control"
              rows="3"
              placeholder="e.g., Payment not found, incorrect amount, etc."
              value={cancelReason}
              onChange={(event) => setCancelReason(event.target.value)}
              required
              style={{ borderRadius: '12px' }}
            />
          </div>
        )}
      </ConfirmationModal>
    </div>
  )
}

export default AdminDeposits
