import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const Deposits = () => {
  const [rows, setRows] = useState([])
  const [form, setForm] = useState({ amount: '', provider: '' })
  const [loading, setLoading] = useState(false)

  const load = async () => {
    const response = await walletService.getMyDeposits({ limit: 30 })
    setRows(getListPayload(response).rows)
  }

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load deposits')))
  }, [])

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      await walletService.createDeposit(form)
      toast.success('Deposit request created')
      setForm({ amount: '', provider: '' })
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to create deposit'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Deposits</h1>
          <p>Create deposit requests and track status.</p>
        </div>
      </div>

      <section className="form-panel mb-4">
        <form onSubmit={submit} className="inline-form">
          <input className="form-control" placeholder="Amount" value={form.amount} onChange={(event) => setForm({ ...form, amount: event.target.value })} required />
          <input className="form-control" placeholder="Provider optional" value={form.provider} onChange={(event) => setForm({ ...form, provider: event.target.value })} />
          <button className="btn btn-primary" disabled={loading}>{loading ? 'Creating...' : 'Create deposit'}</button>
        </form>
      </section>

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
                <td>{row.payment_reference || row.id}</td>
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

export default Deposits
