import { useState } from 'react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'
import { formatMoney } from '../../utils/deposit'

const DepositForm = ({ onClose, onCreated }) => {
  const [form, setForm] = useState({ amount: '', provider: 'qris' })
  const [loading, setLoading] = useState(false)
  const amount = Number(form.amount || 0)
  const belowMinimum = amount > 0 && amount < 10000
  const estimatedFee = form.provider === 'qris' ? amount * 0.05 : 0
  const estimatedPayable = amount + estimatedFee

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      const response = await walletService.createDeposit(form)
      toast.success('Deposit request created')
      onCreated(response.data.data)
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
            <p>Choose QRIS or manual review for wallet topup.</p>
          </div>
          <button className="modal-close" type="button" onClick={onClose} aria-label="Close">
            <i className="bi bi-x-lg"></i>
          </button>
        </div>
        <form onSubmit={submit}>
          <label className="form-label">Payment Method</label>
          <select
            className="form-select"
            value={form.provider}
            onChange={(event) => setForm({ ...form, provider: event.target.value })}
          >
            <option value="qris">QRIS</option>
            <option value="">Manual review</option>
          </select>

          <label className="form-label mt-3">Amount</label>
          <input
            className="form-control"
            type="number"
            min="10000"
            placeholder="Minimum 10000"
            value={form.amount}
            onChange={(event) => setForm({ ...form, amount: event.target.value })}
            required
          />
          {belowMinimum && <div className="form-hint text-danger">Minimum deposit {formatMoney(10000)}</div>}
          {form.provider === 'qris' && (
            <div className="payment-summary mt-3">
              <div><span>Wallet credit</span><strong>{formatMoney(amount)}</strong></div>
              <div><span>Topup fee 5%</span><strong>{formatMoney(estimatedFee)}</strong></div>
              <div><span>Estimated payment</span><strong>{formatMoney(estimatedPayable)} + kode unik</strong></div>
            </div>
          )}
          <div className="toolbar-actions justify-content-end mt-4">
            <button className="btn btn-outline-dark" type="button" onClick={onClose}>Cancel</button>
            <button className="btn btn-primary" disabled={loading || belowMinimum}>{loading ? 'Creating...' : 'Create deposit'}</button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default DepositForm
