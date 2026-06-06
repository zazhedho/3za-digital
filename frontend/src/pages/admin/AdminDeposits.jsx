import { useCallback, useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import TableActionMenu from '../../components/common/TableActionMenu'
import { depositMetadata, depositPayableAmount, depositStatus, depositStatusClass, formatMoney, isQRISDeposit } from '../../utils/deposit'
import { useAuth } from '../../contexts/AuthContext'
import PaginationBar from '../../components/common/PaginationBar'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual deposit'
  if (method === 'payment_gateway') return 'Payment gateway'
  return method || 'Deposit request'
}

const AdminDeposits = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 50 })
  const [loading, setLoading] = useState(false)
  const [confirmAction, setConfirmAction] = useState(null)
  const [confirmLoading, setConfirmLoading] = useState(false)
  const [cancelReason, setCancelReason] = useState('')
  const canApproveDeposit = hasPermission('admin_deposits', 'approve')
  const canCancelDeposit = hasPermission('admin_deposits', 'cancel')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await walletService.getDeposits({ page, limit: 50 })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } finally {
      setLoading(false)
    }
  }, [page])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposits')))
  }, [load])

  const approve = async () => {
    if (!confirmAction?.deposit) return
    setConfirmLoading(true)
    try {
      await walletService.updateDepositStatus(confirmAction.deposit.id, { status: 'paid', amount: confirmAction.deposit.amount, description: 'Approved from frontend' })
      toast.success('Deposit approved')
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
      await walletService.updateDepositStatus(confirmAction.deposit.id, { status: 'cancelled', reason: cancelReason.trim() })
      toast.success('Deposit cancelled')
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

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Admin Deposits</h1>
          <p>Review and approve user deposits.</p>
        </div>
      </div>
      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>User</th>
              <th>Reference</th>
              <th>Amount</th>
              <th>Pay</th>
              <th>Status</th>
              <th>Reason</th>
              <th className="text-end">Action</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => {
              const metadata = depositMetadata(row)
              return (
                <tr key={row.id}>
                  <td>
                    <span className="table-main">
                      <strong>{row.user?.name || row.user?.email || 'User'}</strong>
                      {row.user?.email && row.user?.name && <span className="table-subtext">{row.user.email}</span>}
                    </span>
                  </td>
                  <td>
                    <span className="table-main">
                      <strong>{methodLabel(row.method)}</strong>
                      <span className="table-subtext">{row.payment_reference ? `Ref ${row.payment_reference}` : 'Waiting for reference'}</span>
                    </span>
                  </td>
                  <td className="table-number">{formatMoney(row.amount)}</td>
                  <td className="table-number">{isQRISDeposit(row) ? formatMoney(depositPayableAmount(row)) : '-'}</td>
                  <td><span className={`status-badge ${depositStatusClass(row)} text-capitalize`}>{depositStatus(row)}</span></td>
                  <td><span className="table-subtext">{metadata.cancel_reason || '-'}</span></td>
                  <td className="text-end">
                    <TableActionMenu
                      label="Open deposit actions"
                      items={[
                        { label: 'Detail', to: `/admin/deposits/${row.id}` },
                        { label: 'Approve', hidden: !canApproveDeposit, disabled: row.status !== 'pending', onClick: () => openConfirm('approve', row) },
                        { label: 'Cancel', hidden: !canCancelDeposit, disabled: row.status !== 'pending', danger: true, onClick: () => openConfirm('cancel', row) },
                      ]}
                    />
                  </td>
                </tr>
              )
            })}
            {!rows.length && <tr><td colSpan="7" className="empty-cell">{loading ? 'Loading...' : 'No deposits found'}</td></tr>}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
      <ConfirmationModal
        show={Boolean(confirmAction)}
        title={isCancelAction ? 'Cancel Deposit' : 'Approve Deposit'}
        message={`${isCancelAction ? 'Cancel' : 'Approve'} deposit ${formatMoney(confirmDeposit?.amount)} for ${confirmDeposit?.user?.name || confirmDeposit?.user?.email || 'this user'}?`}
        confirmLabel={isCancelAction ? 'Cancel deposit' : 'Approve'}
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
          <div>
            <label className="form-label">Reason</label>
            <textarea
              className="form-control"
              placeholder="Example: duplicate payment request"
              value={cancelReason}
              onChange={(event) => setCancelReason(event.target.value)}
            />
          </div>
        )}
      </ConfirmationModal>
    </div>
  )
}

export default AdminDeposits
