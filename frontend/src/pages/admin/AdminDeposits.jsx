import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'
import { depositPayableAmount, depositStatus, depositStatusClass, formatMoney, isQRISDeposit } from '../../utils/deposit'

const methodLabel = (method) => {
  if (method === 'manual_admin') return 'Manual deposit'
  if (method === 'payment_gateway') return 'Payment gateway'
  return method || 'Deposit request'
}

const AdminDeposits = () => {
  const [rows, setRows] = useState([])

  const load = async () => {
    const response = await walletService.getDeposits({ limit: 50 })
    setRows(getListPayload(response).rows)
  }

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposits')))
  }, [])

  const approve = async (row) => {
    try {
      await walletService.approveDeposit(row.id, { amount: row.amount, description: 'Approved from frontend' })
      toast.success('Deposit approved')
      load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Approve failed'))
    }
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
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><span className="table-main"><strong>{row.user?.name || row.user?.email || 'User'}</strong></span></td>
                <td>
                  <span className="table-main">
                    <strong>{methodLabel(row.method)}</strong>
                    <span className="table-subtext">{row.payment_reference ? `Ref ${row.payment_reference}` : 'Waiting for reference'}</span>
                  </span>
                </td>
                <td className="table-number">{formatMoney(row.amount)}</td>
                <td className="table-number">{isQRISDeposit(row) ? formatMoney(depositPayableAmount(row)) : '-'}</td>
                <td><span className={`status-badge ${depositStatusClass(row)} text-capitalize`}>{depositStatus(row)}</span></td>
                <td className="text-end">
                  <span className="table-actions">
                    <Link className="btn btn-sm btn-outline-dark" to={`/admin/deposits/${row.id}`}>Detail</Link>
                    <button className="btn btn-sm btn-primary" disabled={row.status !== 'pending'} onClick={() => approve(row)}>Approve</button>
                  </span>
                </td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="6" className="empty-cell">No deposits found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default AdminDeposits
