import { useState } from 'react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'

const DepositForm = ({ onClose, onCreated }) => {
  const [form, setForm] = useState({ amount: '', provider: '' })
  const [loading, setLoading] = useState(false)

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      await walletService.createDeposit(form)
      toast.success('Deposit request created')
      onCreated()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to create deposit'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-backdrop-custom" role="dialog" aria-modal="true">
      <div className="modal-panel deposit-modal">
        <div className="modal-heading">
          <div>
            <h5>Create Deposit</h5>
            <p>Submit a manual deposit request for wallet balance.</p>
          </div>
          <button className="modal-close" type="button" onClick={onClose} aria-label="Close">
            <i className="bi bi-x-lg"></i>
          </button>
        </div>
        <form onSubmit={submit}>
          <label className="form-label">Amount</label>
          <input
            className="form-control"
            placeholder="Example: 100000"
            value={form.amount}
            onChange={(event) => setForm({ ...form, amount: event.target.value })}
            required
          />
          <label className="form-label mt-3">Provider</label>
          <input
            className="form-control"
            placeholder="Optional"
            value={form.provider}
            onChange={(event) => setForm({ ...form, provider: event.target.value })}
          />
          <div className="toolbar-actions justify-content-end mt-4">
            <button className="btn btn-outline-dark" type="button" onClick={onClose}>Cancel</button>
            <button className="btn btn-primary" disabled={loading}>{loading ? 'Creating...' : 'Create deposit'}</button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default DepositForm
