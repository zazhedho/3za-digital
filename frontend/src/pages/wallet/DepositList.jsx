import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import DepositForm from './DepositForm'
import { depositPayableAmount, depositStatus, depositStatusClass, formatMoney, isQRISDeposit } from '../../utils/deposit'

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
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Reference</th>
              <th>Amount</th>
              <th>Pay</th>
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
                  <span className="table-main">
                    <strong>{methodLabel(row.method)}</strong>
                    <span className="table-subtext">{row.payment_reference ? `Ref ${row.payment_reference}` : 'Waiting for reference'}</span>
                  </span>
                </td>
                <td className="table-number">{formatMoney(row.amount)}</td>
                <td className="table-number">{isQRISDeposit(row) ? formatMoney(depositPayableAmount(row)) : '-'}</td>
                <td className="text-capitalize table-nowrap">{row.provider || '-'}</td>
                <td><span className={`status-badge ${depositStatusClass(row)} text-capitalize`}>{depositStatus(row)}</span></td>
                <td className="table-date">{row.created_at ? new Date(row.created_at).toLocaleString('id-ID') : '-'}</td>
                <td className="text-end"><span className="table-actions"><Link className="btn btn-sm btn-outline-dark" to={`/deposits/${row.id}`}>Detail</Link></span></td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="7" className="empty-cell">No deposits found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default DepositList
