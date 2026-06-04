import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import DepositForm from './DepositForm'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual deposit'
  if (method === 'payment_gateway') return 'Payment gateway'
  return method || 'Deposit request'
}

const DepositList = () => {
  const [rows, setRows] = useState([])
  const [showCreate, setShowCreate] = useState(false)

  const load = async () => {
    const response = await walletService.getMyDeposits({ limit: 30 })
    setRows(getListPayload(response).rows)
  }

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposits')))
  }, [])

  const handleCreated = async () => {
    setShowCreate(false)
    await load()
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Deposits</h1>
          <p>Create deposit requests and track status.</p>
        </div>
        <button className="btn btn-primary" type="button" onClick={() => setShowCreate(true)}>
          <i className="bi bi-plus-lg me-2"></i>Create deposit
        </button>
      </div>

      {showCreate && (
        <DepositForm onClose={() => setShowCreate(false)} onCreated={handleCreated} />
      )}

      <section className="table-panel">
        <table className="table align-middle">
          <thead>
            <tr>
              <th>Reference</th>
              <th>Amount</th>
              <th>Provider</th>
              <th>Status</th>
              <th>Date</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <strong>{methodLabel(row.method)}</strong>
                  <div className="text-muted small">{row.payment_reference ? `Ref ${row.payment_reference}` : 'Waiting for reference'}</div>
                </td>
                <td>{formatMoney(row.amount)}</td>
                <td>{row.provider || '-'}</td>
                <td><span className="badge rounded-pill text-bg-light text-capitalize">{row.status}</span></td>
                <td>{row.created_at ? new Date(row.created_at).toLocaleString('id-ID') : '-'}</td>
                <td className="text-end"><Link className="btn btn-sm btn-outline-dark" to={`/deposits/${row.id}`}>Detail</Link></td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="6" className="text-center text-muted py-5">No deposits found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default DepositList
