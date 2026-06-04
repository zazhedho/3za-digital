import { useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage, getListPayload } from '../../services/api'

const AdminWallets = () => {
  const [rows, setRows] = useState([])
  const [activeWallet, setActiveWallet] = useState(null)
  const [form, setForm] = useState({ amount: '', direction: 'credit', description: '' })

  const load = async () => {
    const response = await walletService.getWallets({ limit: 50 })
    setRows(getListPayload(response).rows)
  }

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load wallets')))
  }, [])

  const adjust = async (event) => {
    event.preventDefault()
    try {
      await walletService.adminAdjust(activeWallet.user_id, form)
      toast.success('Wallet adjusted')
      setActiveWallet(null)
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Adjustment failed'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Admin Wallets</h1>
          <p>Wallet balances and manual adjustments.</p>
        </div>
      </div>

      {activeWallet && (
        <section className="form-panel mb-4">
          <form onSubmit={adjust} className="inline-form">
            <strong>Adjust {activeWallet.user?.name || activeWallet.user_id}</strong>
            <input className="form-control" placeholder="Amount" value={form.amount} onChange={(event) => setForm({ ...form, amount: event.target.value })} required />
            <select className="form-select" value={form.direction} onChange={(event) => setForm({ ...form, direction: event.target.value })}>
              <option value="credit">Credit</option>
              <option value="debit">Debit</option>
            </select>
            <input className="form-control" placeholder="Description" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} required />
            <button className="btn btn-primary">Save</button>
            <button className="btn btn-outline-secondary" type="button" onClick={() => setActiveWallet(null)}>Cancel</button>
          </form>
        </section>
      )}

      <section className="table-panel">
        <table className="table align-middle">
          <thead>
            <tr>
              <th>User</th>
              <th>Balance</th>
              <th>Currency</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>{row.user?.name || row.user_id}</td>
                <td>{row.balance}</td>
                <td>{row.currency}</td>
                <td>{row.is_active ? 'Active' : 'Inactive'}</td>
                <td className="text-end"><button className="btn btn-sm btn-outline-dark" onClick={() => setActiveWallet(row)}>Adjust</button></td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="5" className="text-center text-muted py-5">No wallets found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default AdminWallets
