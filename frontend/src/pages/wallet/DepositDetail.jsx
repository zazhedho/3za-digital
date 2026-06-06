import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'
import QRISPaymentBox from '../../components/common/QRISPaymentBox'
import { depositMetadata, depositPayableAmount, depositProviderLabel, depositStatus, depositStatusClass, formatMoney, isQRISDeposit, qrisImageURL, qrisString } from '../../utils/deposit'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual deposit'
  if (method === 'payment_gateway') return 'Payment gateway'
  return method || '-'
}

const DepositDetail = () => {
  const { id } = useParams()
  const [deposit, setDeposit] = useState(null)
  const metadata = depositMetadata(deposit)
  const qris = isQRISDeposit(deposit)
  const qrisImage = qrisImageURL(deposit)
  const qrisPayload = qrisString(deposit)

  useEffect(() => {
    walletService.getMyDepositById(id)
      .then((response) => setDeposit(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposit')))
  }, [id])

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Deposit Detail</h1><p>{deposit ? `${methodLabel(deposit.method)} - ${depositStatus(deposit)}` : 'Loading deposit'}</p></div>
        <BackButton fallback="/deposits" />
      </div>
      <section className="panel">
        <div className="detail-grid detail-grid-compact">
          <span>ID</span><strong>{deposit?.id || '-'}</strong>
          <span>Wallet Credit</span><strong>{formatMoney(deposit?.amount)}</strong>
          {qris && <><span>Topup Fee</span><strong>{formatMoney(metadata.fee_amount)}</strong></>}
          {qris && <><span>Unique Code</span><strong>{formatMoney(metadata.unique_code_amount)}</strong></>}
          {qris && <><span>Total Payment</span><strong>{formatMoney(depositPayableAmount(deposit))}</strong></>}
          <span>Status</span><div className="detail-value"><span className={`status-badge status-badge-detail ${depositStatusClass(deposit)} text-capitalize`}>{depositStatus(deposit)}</span></div>
          <span>Method</span><strong>{methodLabel(deposit?.method)}</strong>
          <span>Provider</span><strong>{depositProviderLabel(deposit)}</strong>
          {deposit?.payment_reference && <><span>Reference</span><strong>{deposit.payment_reference}</strong></>}
          {qris && metadata.qris_merchant_name && <><span>Merchant</span><strong>{metadata.qris_merchant_name}</strong></>}
          {metadata.cancel_reason && <><span>Cancel Reason</span><strong>{metadata.cancel_reason}</strong></>}
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
    </div>
  )
}

export default DepositDetail
