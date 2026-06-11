import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'
import QRISPaymentBox from '../../components/common/QRISPaymentBox'
import { depositMetadata, depositPayableAmount, depositProviderLabel, depositStatus, depositStatusClass, formatMoney, isQRISDeposit, qrisImageURL, qrisString } from '../../utils/deposit'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual'
  if (method === 'payment_gateway') return 'Payment Gateway'
  return method || '-'
}

const DepositDetail = () => {
  const { id } = useParams()
  const [deposit, setDeposit] = useState(null)
  const [loading, setLoading] = useState(true)
  const metadata = depositMetadata(deposit)
  const qris = isQRISDeposit(deposit)
  const qrisImage = qrisImageURL(deposit)
  const qrisPayload = qrisString(deposit)

  useEffect(() => {
    setLoading(true)
    walletService.getMyDepositById(id)
      .then((response) => setDeposit(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposit')))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Deposit Detail</h1>
          <p>Transaction reference and payment info</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/deposits" />
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
            <i className="bi bi-wallet2"></i> {methodLabel(deposit?.method)}
          </div>
          <h2 className="luxe-hero-title">{formatMoney(deposit?.amount)}</h2>
          <p className="luxe-hero-subtitle">Credit to wallet balance</p>
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
            <h3><i className="bi bi-info-circle"></i> Transaction Info</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Deposit ID</span>
                <span className="luxe-value-code">{deposit?.id}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Provider</span>
                <span className="luxe-value">{depositProviderLabel(deposit)}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Reference</span>
                <span className="luxe-value">{deposit?.payment_reference || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Created At</span>
                <span className="luxe-value">{deposit?.created_at ? new Date(deposit.created_at).toLocaleString('id-ID') : '-'}</span>
              </div>
            </div>
          </div>
        </section>

        {qris ? (
          <section className="luxe-detail-card">
            <div className="luxe-card-header">
              <h3><i className="bi bi-qr-code-scan"></i> Payment Summary</h3>
            </div>
            <div className="luxe-card-body">
              <div className="luxe-grid">
                <div className="luxe-item">
                  <span className="luxe-label">Topup Amount</span>
                  <span className="luxe-value">{formatMoney(deposit?.amount)}</span>
                </div>
                <div className="luxe-item">
                  <span className="luxe-label">Service Fee</span>
                  <span className="luxe-value">{formatMoney(metadata.fee_amount)}</span>
                </div>
                <div className="luxe-item">
                  <span className="luxe-label">Unique Code</span>
                  <span className="luxe-value">{formatMoney(metadata.unique_code_amount)}</span>
                </div>
                <div className="luxe-item">
                  <span className="luxe-label">Total Payable</span>
                  <span className="luxe-value-strong text-primary">{formatMoney(depositPayableAmount(deposit))}</span>
                </div>
              </div>
              
              <div className="mt-4">
                <QRISPaymentBox
                  amount={formatMoney(depositPayableAmount(deposit))}
                  description={`Please pay the exact amount to confirm your deposit.`}
                  image={qrisImage}
                  payload={qrisPayload}
                />
              </div>
            </div>
          </section>
        ) : (
          <section className="luxe-detail-card">
            <div className="luxe-card-header">
              <h3><i className="bi bi-patch-check"></i> Status & Notes</h3>
            </div>
            <div className="luxe-card-body">
               <div className="luxe-grid">
                  <div className="luxe-item">
                    <span className="luxe-label">Current Status</span>
                    <span className="luxe-value text-capitalize">{depositStatus(deposit)}</span>
                  </div>
                  {metadata.cancel_reason && (
                    <div className="luxe-item">
                      <span className="luxe-label">Cancellation Reason</span>
                      <span className="luxe-value text-danger">{metadata.cancel_reason}</span>
                    </div>
                  )}
                  {deposit?.paid_at && (
                    <div className="luxe-item">
                      <span className="luxe-label">Paid At</span>
                      <span className="luxe-value">{new Date(deposit.paid_at).toLocaleString('id-ID')}</span>
                    </div>
                  )}
               </div>
               
               {deposit?.status === 'pending' && (
                 <div className="auth-alert mt-4">
                    <i className="bi bi-info-circle me-2"></i>
                    For manual review, please contact our support team after making the payment.
                 </div>
               )}
            </div>
          </section>
        )}
      </div>
    </div>
  )
}

export default DepositDetail
