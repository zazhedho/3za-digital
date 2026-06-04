import { useCallback, useEffect, useState } from 'react'
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

const AdminDepositDetail = () => {
  const { id } = useParams()
  const [deposit, setDeposit] = useState(null)

  const load = useCallback(async () => {
    const response = await walletService.getDepositById(id)
    setDeposit(response.data.data)
  }, [id])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposit')))
  }, [load])

  const approve = async () => {
    try {
      await walletService.approveDeposit(id, { amount: deposit.amount, description: 'Approved from frontend' })
      toast.success('Deposit approved')
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Approve failed'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Admin Deposit Detail</h1><p>{deposit ? `${methodLabel(deposit.method)} - ${deposit.status || 'pending'}` : 'Loading deposit'}</p></div>
        <div className="toolbar-actions">
          <BackButton fallback="/admin/deposits" />
          <button className="btn btn-primary" disabled={deposit?.status !== 'pending'} onClick={approve}>Approve</button>
        </div>
      </div>
      <section className="panel">
        <div className="detail-grid">
          <span>ID</span><strong>{deposit?.id || '-'}</strong>
          <span>User</span><strong>{deposit?.user?.name || deposit?.user?.email || '-'}</strong>
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

export default AdminDepositDetail
