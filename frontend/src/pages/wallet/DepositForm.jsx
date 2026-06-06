import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { QRCodeSVG } from 'qrcode.react'
import { toast } from 'react-toastify'
import walletService from '../../services/walletService'
import { getErrorMessage } from '../../services/api'
import { depositPayableAmount, formatMoney, isQRISDeposit, qrisImageURL, qrisString } from '../../utils/deposit'

const DepositForm = ({ onClose, onCreated }) => {
  const [form, setForm] = useState({ amount: '', provider: 'qris' })
  const [settings, setSettings] = useState({ minimumAmount: 10000, qrisFeePercent: '', qrisStaticImageURL: '' })
  const [settingsLoading, setSettingsLoading] = useState(true)
  const [createdDeposit, setCreatedDeposit] = useState(null)
  const [loading, setLoading] = useState(false)
  const amount = Number(form.amount || 0)
  const minimumAmount = Number(settings.minimumAmount || 10000)
  const feePercent = Number(settings.qrisFeePercent)
  const qrisFeeKnown = Number.isFinite(feePercent)
  const belowMinimum = amount > 0 && amount < minimumAmount
  const isQRISProvider = ['qris', 'qrisly'].includes(form.provider)
  const staticQRISReady = Boolean(settings.qrisStaticImageURL)
  const estimatedFee = isQRISProvider && qrisFeeKnown ? amount * (feePercent / 100) : 0
  const estimatedPayable = amount + estimatedFee
  const createdIsQRIS = isQRISDeposit(createdDeposit)
  const createdQRISImage = qrisImageURL(createdDeposit)
  const createdQRISPayload = qrisString(createdDeposit)

  useEffect(() => {
    walletService.getDepositSettings()
      .then((response) => {
        const data = response.data.data || {}
        const staticImageURL = data.qris_static_image_url || ''
        setSettings({
          minimumAmount: Number(data.minimum_amount || 10000),
          qrisFeePercent: data.qris_fee_percent ?? '',
          qrisStaticImageURL: staticImageURL,
        })
        if (!staticImageURL) setForm((current) => (current.provider === 'qris' ? { ...current, provider: 'qrisly' } : current))
      })
      .catch(() => {
        setSettings((current) => ({ ...current, qrisFeePercent: '', qrisStaticImageURL: '' }))
        setForm((current) => (current.provider === 'qris' ? { ...current, provider: 'qrisly' } : current))
      })
      .finally(() => setSettingsLoading(false))
  }, [])

  const submit = async (event) => {
    event.preventDefault()
    setLoading(true)
    try {
      const response = await walletService.createDeposit(form)
      toast.success('Deposit request created')
      const deposit = response.data.data
      setCreatedDeposit(deposit)
      onCreated(deposit)
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
            <h5>{createdDeposit ? 'Scan QRIS' : 'Create Deposit'}</h5>
            <p>{createdDeposit ? 'Pay the exact amount shown.' : 'Choose static QRIS, dynamic QRIS, or manual review for wallet topup.'}</p>
          </div>
          <button className="modal-close" type="button" onClick={onClose} aria-label="Close">
            <i className="bi bi-x-lg"></i>
          </button>
        </div>
        {createdDeposit ? (
          <div className="deposit-created">
            {createdIsQRIS ? (
              <>
                <div className="qris-modal-code">
                  {createdQRISImage ? (
                    <img src={createdQRISImage} alt="QRIS payment code" />
                  ) : createdQRISPayload ? (
                    <QRCodeSVG value={createdQRISPayload} size={260} level="M" includeMargin />
                  ) : (
                    <div className="qris-placeholder">QRIS is not available yet</div>
                  )}
                </div>
                <div className="deposit-created-total">
                  <span>Total Payment</span>
                  <strong>{formatMoney(depositPayableAmount(createdDeposit))}</strong>
                </div>
              </>
            ) : (
              <div className="payment-summary">
                <div><span>Deposit request</span><strong>{formatMoney(createdDeposit.amount)}</strong></div>
                <div><span>Status</span><strong>{createdDeposit.status || '-'}</strong></div>
              </div>
            )}
            <div className="toolbar-actions justify-content-end mt-4">
              <button className="btn btn-outline-dark" type="button" onClick={onClose}>Done</button>
              {createdDeposit?.id && <Link className="btn btn-primary" to={`/deposits/${createdDeposit.id}`} onClick={onClose}>Detail</Link>}
            </div>
          </div>
        ) : (
        <form onSubmit={submit}>
          <label className="form-label">Payment Method</label>
          <select
            className="form-select"
            value={form.provider}
            onChange={(event) => setForm({ ...form, provider: event.target.value })}
          >
            <option value="qris" disabled={!staticQRISReady}>Static QRIS</option>
            <option value="qrisly">Dynamic QRIS</option>
            <option value="">Manual review</option>
          </select>
          {form.provider === 'qris' && !staticQRISReady && <div className="form-hint text-danger">Static QRIS image URL is not configured.</div>}

          <label className="form-label mt-3">Amount</label>
          <input
            className="form-control"
            type="number"
            min={minimumAmount}
            placeholder={`Minimum ${minimumAmount}`}
            value={form.amount}
            onChange={(event) => setForm({ ...form, amount: event.target.value })}
            required
          />
          {belowMinimum && <div className="form-hint text-danger">Minimum deposit {formatMoney(minimumAmount)}</div>}
          {isQRISProvider && (
            <div className="payment-summary mt-3">
              <div><span>Wallet credit</span><strong>{formatMoney(amount)}</strong></div>
              <div>
                <span>Topup Fee</span>
                <strong>{settingsLoading ? 'Loading...' : (qrisFeeKnown ? formatMoney(estimatedFee) : 'Calculated after create')}</strong>
              </div>
              <div>
                <span>Estimated payment</span>
                <strong>{qrisFeeKnown ? `${formatMoney(estimatedPayable)} + unique code` : 'Calculated after create'}</strong>
              </div>
            </div>
          )}
          <div className="toolbar-actions justify-content-end mt-4">
            <button className="btn btn-outline-dark" type="button" onClick={onClose}>Cancel</button>
            <button className="btn btn-primary" disabled={loading || belowMinimum || (form.provider === 'qris' && !staticQRISReady)}>{loading ? 'Creating...' : 'Create deposit'}</button>
          </div>
        </form>
        )}
      </div>
    </div>
  )
}

export default DepositForm
