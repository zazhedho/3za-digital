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
            <h5>{createdDeposit ? (createdIsQRIS ? 'Scan QRIS' : 'Deposit Requested') : 'Create Deposit'}</h5>
            <p>{createdDeposit ? (createdIsQRIS ? 'Pay the exact amount shown.' : 'Please wait for admin approval.') : 'Choose static QRIS, dynamic QRIS, or manual review for wallet topup.'}</p>
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
              <div className="deposit-manual-success">
                <div className="manual-success-icon">
                  <i className="bi bi-headset"></i>
                </div>
                <h3>Action Required</h3>
                <p>To complete your {form.provider === 'qris' ? 'Static QRIS' : 'Manual'} deposit, please transfer the exact amount and <strong>contact our Admin</strong> via the Support Center.</p>
                <div className="manual-success-details">
                  <div className="detail-row">
                    <span>Amount</span>
                    <strong>{formatMoney(createdDeposit.amount)}</strong>
                  </div>
                  <div className="detail-row">
                    <span>Status</span>
                    <span className="status-badge warning">{createdDeposit.status || 'Pending'}</span>
                  </div>
                </div>
              </div>
            )}
            <div className="toolbar-actions justify-content-end mt-4">
              <button className="btn btn-outline-dark" type="button" onClick={onClose}>Done</button>
              {createdDeposit?.id && <Link className="btn btn-primary" to={`/deposits/${createdDeposit.id}`} onClick={onClose}>Detail</Link>}
            </div>
          </div>
        ) : (
        <form onSubmit={submit} className="deposit-form-modern">
          <div className="deposit-input-group">
            <label>Payment Method</label>
            <div className="deposit-select-wrapper">
              <i className="bi bi-wallet2"></i>
              <select
                value={form.provider}
                onChange={(event) => setForm({ ...form, provider: event.target.value })}
              >
                <option value="qris" disabled={!settingsLoading && !staticQRISReady}>Static QRIS</option>
                <option value="qrisly">Dynamic QRIS</option>
                <option value="">Manual review</option>
              </select>
            </div>
            {form.provider === 'qris' && !staticQRISReady && !settingsLoading && <div className="form-hint text-danger mt-1"><i className="bi bi-exclamation-circle"></i> Static QRIS image URL is not configured.</div>}
          </div>

          <div className="deposit-input-group">
            <label>Amount</label>
            <div className="deposit-amount-wrapper">
              <span className="deposit-amount-currency">Rp</span>
              <input
                type="number"
                min={minimumAmount}
                placeholder={minimumAmount.toLocaleString('id-ID')}
                value={form.amount}
                onChange={(event) => setForm({ ...form, amount: event.target.value })}
                required
              />
            </div>
            {belowMinimum && <div className="form-hint text-danger mt-1"><i className="bi bi-exclamation-circle"></i> Minimum deposit {formatMoney(minimumAmount)}</div>}
          </div>

          {isQRISProvider && (
            <div className="deposit-fee-card">
              <div className="deposit-fee-row">
                <span>Wallet credit</span>
                <strong>{formatMoney(amount)}</strong>
              </div>
              <div className="deposit-fee-row">
                <span>Topup Fee</span>
                <strong>{settingsLoading ? 'Loading...' : (qrisFeeKnown ? formatMoney(estimatedFee) : 'TBD')}</strong>
              </div>
              <div className="deposit-fee-row total">
                <span>Estimated payment</span>
                <strong>{qrisFeeKnown ? `${formatMoney(estimatedPayable)} + code` : 'TBD'}</strong>
              </div>
            </div>
          )}
          
          <div className="deposit-modal-actions">
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
