import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual deposit'
  if (method === 'payment_gateway') return 'Payment gateway'
  return method || '-'
}

const DepositDetail = () => {
  const { id } = useParams()
  const [deposit, setDeposit] = useState(null)

  useEffect(() => {
    walletService.getMyDepositById(id)
      .then((response) => setDeposit(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposit')))
  }, [id])

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Deposit Detail</h1><p>{deposit ? `${methodLabel(deposit.method)} - ${deposit.status || 'pending'}` : 'Loading deposit'}</p></div>
        <BackButton fallback="/deposits" />
      </div>
      <section className="panel">
        <div className="detail-grid">
          <span>ID</span><strong>{deposit?.id || '-'}</strong>
          <span>Amount</span><strong>{deposit?.amount || '-'}</strong>
          <span>Status</span><strong>{deposit?.status || '-'}</strong>
          <span>Method</span><strong>{methodLabel(deposit?.method)}</strong>
          <span>Provider</span><strong>{deposit?.provider || '-'}</strong>
          {deposit?.payment_reference && <><span>Reference</span><strong>{deposit.payment_reference}</strong></>}
          <span>Created</span><strong>{deposit?.created_at ? new Date(deposit.created_at).toLocaleString('id-ID') : '-'}</strong>
        </div>
      </section>
    </div>
  )
}

export default DepositDetail
