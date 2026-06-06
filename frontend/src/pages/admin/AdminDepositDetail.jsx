import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'
import QRISPaymentBox from '../../components/common/QRISPaymentBox'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import { depositMetadata, depositPayableAmount, depositProviderLabel, depositStatus, depositStatusClass, formatMoney, isQRISDeposit, qrisImageURL, qrisString } from '../../utils/deposit'
import { useAuth } from '../../contexts/AuthContext'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual deposit'
  if (method === 'payment_gateway') return 'Payment gateway'
  return method || '-'
}

const AdminDepositDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [deposit, setDeposit] = useState(null)
  const [confirmAction, setConfirmAction] = useState(null)
  const [confirmLoading, setConfirmLoading] = useState(false)
  const [cancelReason, setCancelReason] = useState('')
  const canApproveDeposit = hasPermission('admin_deposits', 'approve')
  const canCancelDeposit = hasPermission('admin_deposits', 'cancel')
  const metadata = depositMetadata(deposit)
  const qris = isQRISDeposit(deposit)
  const qrisImage = qrisImageURL(deposit)
  const qrisPayload = qrisString(deposit)

  const load = useCallback(async () => {
    const response = await walletService.getDepositById(id)
    setDeposit(response.data.data)
  }, [id])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposit')))
  }, [load])

  const approve = async () => {
    setConfirmLoading(true)
    try {
      await walletService.updateDepositStatus(id, { status: 'paid', amount: deposit.amount, description: 'Approved from frontend' })
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
    setConfirmLoading(true)
    try {
      await walletService.updateDepositStatus(id, { status: 'cancelled', reason: cancelReason.trim() })
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

  const isCancelAction = confirmAction === 'cancel'
  const openConfirm = (type) => {
    setCancelReason('')
    setConfirmAction(type)
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Admin Deposit Detail</h1><p>{deposit ? `${methodLabel(deposit.method)} - ${depositStatus(deposit)}` : 'Loading deposit'}</p></div>
        <div className="toolbar-actions">
          <BackButton fallback="/admin/deposits" />
        </div>
      </div>
      <section className="panel">
        <div className="detail-grid detail-grid-compact">
          <span>ID</span><strong>{deposit?.id || '-'}</strong>
          <span>User</span><strong>{deposit?.user?.name || deposit?.user?.email || '-'}</strong>
          <span>Wallet Credit</span><strong>{formatMoney(deposit?.amount)}</strong>
          {qris && <><span>Topup Fee</span><strong>{formatMoney(metadata.fee_amount)}</strong></>}
          {qris && <><span>Unique Code</span><strong>{formatMoney(metadata.unique_code_amount)}</strong></>}
          {qris && <><span>Total Payment</span><strong>{formatMoney(depositPayableAmount(deposit))}</strong></>}
          <span>Status</span><div className="detail-value"><span className={`status-badge status-badge-detail ${depositStatusClass(deposit)} text-capitalize`}>{depositStatus(deposit)}</span></div>
          <span>Action</span>
          <strong className="detail-action-cell">
            <div className="detail-actions">
              {canApproveDeposit && <button className="btn btn-sm btn-primary" disabled={deposit?.status !== 'pending'} onClick={() => openConfirm('approve')}>Approve</button>}
              {canCancelDeposit && <button className="btn btn-sm btn-outline-danger" disabled={deposit?.status !== 'pending'} onClick={() => openConfirm('cancel')}>Cancel</button>}
              {!canApproveDeposit && !canCancelDeposit && <span className="table-subtext">No action available</span>}
            </div>
          </strong>
          <span>Method</span><strong>{methodLabel(deposit?.method)}</strong>
          <span>Provider</span><strong>{depositProviderLabel(deposit)}</strong>
          {deposit?.payment_reference && <><span>Reference</span><strong>{deposit.payment_reference}</strong></>}
          {qris && metadata.qris_merchant_name && <><span>Merchant</span><strong>{metadata.qris_merchant_name}</strong></>}
          {metadata.cancel_reason && <><span>Cancel Reason</span><strong>{metadata.cancel_reason}</strong></>}
          {metadata.cancelled_at && <><span>Cancelled At</span><strong>{new Date(metadata.cancelled_at).toLocaleString('id-ID')}</strong></>}
          <span>Created</span><strong>{deposit?.created_at ? new Date(deposit.created_at).toLocaleString('id-ID') : '-'}</strong>
        </div>
        {qris && (
          <QRISPaymentBox
            amount={formatMoney(depositPayableAmount(deposit))}
            description={`Wallet credit ${formatMoney(deposit?.amount)} after payment confirmation.`}
            image={qrisImage}
            payload={qrisPayload}
          />
        )}
      </section>
      <ConfirmationModal
        show={Boolean(confirmAction)}
        title={isCancelAction ? 'Cancel Deposit' : 'Approve Deposit'}
        message={`${isCancelAction ? 'Cancel' : 'Approve'} this deposit for ${deposit?.user?.name || deposit?.user?.email || 'this user'}?`}
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

export default AdminDepositDetail
