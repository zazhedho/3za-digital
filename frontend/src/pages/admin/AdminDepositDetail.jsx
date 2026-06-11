import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import { depositMetadata, depositPayableAmount, depositProviderLabel, depositStatus, depositStatusClass, formatMoney, isQRISDeposit, qrisImageURL, qrisString } from '../../utils/deposit'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual'
  if (method === 'payment_gateway') return 'Payment Gateway'
  return method || '-'
}

const AdminDepositDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [deposit, setDeposit] = useState(null)
  const [loading, setLoading] = useState(true)
  const [confirmAction, setConfirmAction] = useState(null) // 'approve' or 'cancel'
  const [confirmLoading, setConfirmLoading] = useState(false)
  const [cancelReason, setCancelReason] = useState('')

  const metadata = depositMetadata(deposit)
  const qris = isQRISDeposit(deposit)
  const canApproveDeposit = hasPermission('admin_deposits', 'approve')
  const canCancelDeposit = hasPermission('admin_deposits', 'cancel')

  const fetchData = async () => {
    try {
      const response = await walletService.getDepositById(id)
      setDeposit(response.data.data)
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load deposit'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [id])

  const handleConfirm = async () => {
    setConfirmLoading(true)
    try {
      if (confirmAction === 'approve') {
        await walletService.approveDeposit(id, { amount: deposit.amount })
        toast.success('Deposit approved successfully')
      } else if (confirmAction === 'cancel') {
        if (!cancelReason.trim()) {
          toast.error('Cancellation reason is required')
          return
        }
        await walletService.cancelDeposit(id, { reason: cancelReason })
        toast.success('Deposit cancelled')
      }
      setConfirmAction(null)
      fetchData()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Action failed'))
    } finally {
      setConfirmLoading(false)
    }
  }

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Admin Deposit Detail</h1>
          <p>Review and process user deposit request</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/admin/deposits" />
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
            <i className="bi bi-shield-lock"></i> Admin Review - {methodLabel(deposit?.method)}
          </div>
          <h2 className="luxe-hero-title">{formatMoney(deposit?.amount)}</h2>
          <p className="luxe-hero-subtitle">Requester: <strong>{deposit?.user?.name || deposit?.user?.email}</strong></p>
        </div>
        <div className="luxe-hero-badge">
          <span className={`status-badge ${depositStatusClass(deposit)}`}>
            {depositStatus(deposit)}
          </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-info-circle"></i> Request Details</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Transaction ID</span>
                <span className="luxe-value-code">{deposit?.id}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">User Email</span>
                <span className="luxe-value">{deposit?.user?.email || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Payment Provider</span>
                <span className="luxe-value">{depositProviderLabel(deposit)}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Reference Code</span>
                <span className="luxe-value">{deposit?.payment_reference || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Created At</span>
                <span className="luxe-value">{new Date(deposit?.created_at).toLocaleString('id-ID')}</span>
              </div>
              {deposit?.paid_at && (
                <div className="luxe-item">
                   <span className="luxe-label">Settled At</span>
                   <span className="luxe-value">{new Date(deposit.paid_at).toLocaleString('id-ID')}</span>
                </div>
              )}
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-gear-wide-connected"></i> Admin Actions</h3>
          </div>
          <div className="luxe-card-body">
             <div className="luxe-grid">
                <div className="luxe-item">
                   <span className="luxe-label">Current Status</span>
                   <span className="luxe-value text-capitalize">{depositStatus(deposit)}</span>
                </div>
                <div className="luxe-item">
                   <span className="luxe-label">Verification</span>
                   <span className="luxe-value">
                      {deposit?.status === 'pending' ? (
                        <span className="text-warning"><i className="bi bi-hourglass-split"></i> Awaiting Approval</span>
                      ) : (
                        <span className="text-success"><i className="bi bi-check-all"></i> Processed</span>
                      )}
                   </span>
                </div>
             </div>

             {deposit?.status === 'pending' && (
               <div className="toolbar-actions mt-4 pt-4 border-top">
                  {canApproveDeposit && (
                    <button className="btn btn-primary d-flex align-items-center gap-2" onClick={() => setConfirmAction('approve')}>
                       <i className="bi bi-check-circle"></i> Approve Deposit
                    </button>
                  )}
                  {canCancelDeposit && (
                    <button className="btn btn-outline-danger d-flex align-items-center gap-2" onClick={() => setConfirmAction('cancel')}>
                       <i className="bi bi-x-circle"></i> Reject / Cancel
                    </button>
                  )}
               </div>
             )}
             
             {metadata.cancel_reason && (
                <div className="auth-alert mt-4">
                   <strong>Reason for rejection:</strong>
                   <p className="mb-0">{metadata.cancel_reason}</p>
                </div>
             )}
          </div>
        </section>
      </div>

      <ConfirmationModal
        show={Boolean(confirmAction)}
        title={confirmAction === 'approve' ? 'Approve Deposit' : 'Reject Deposit'}
        message={confirmAction === 'approve' ? `Are you sure you want to approve this deposit of ${formatMoney(deposit?.amount)}? This will credit the user wallet immediately.` : 'Please provide a reason for rejecting this deposit.'}
        onConfirm={handleConfirm}
        onCancel={() => setConfirmAction(null)}
        loading={confirmLoading}
        confirmLabel={confirmAction === 'approve' ? 'Approve Now' : 'Reject Now'}
        confirmClassName={confirmAction === 'approve' ? 'btn-primary' : 'btn-danger'}
      >
        {confirmAction === 'cancel' && (
          <div className="mt-3">
            <label className="form-label">Reason</label>
            <textarea
              className="form-control"
              rows="3"
              placeholder="E.g. Payment not found, invalid proof, etc."
              value={cancelReason}
              onChange={(e) => setCancelReason(e.target.value)}
              required
            ></textarea>
          </div>
        )}
      </ConfirmationModal>
    </div>
  )
}

export default AdminDepositDetail
