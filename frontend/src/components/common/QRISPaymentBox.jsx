import { useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'

const QRISVisual = ({ image, payload }) => {
  if (image) return <img src={image} alt="QRIS payment code" />
  if (payload) return (
    <QRCodeSVG 
      value={payload} 
      size={260} 
      level="M" 
      includeMargin 
      style={{ width: '100%', height: 'auto', maxWidth: '260px' }}
    />
  )
  return <div className="qris-placeholder">QRIS is not available yet</div>
}

const QRISPaymentBox = ({ amount, description, image, payload }) => {
  const [open, setOpen] = useState(false)
  const canOpen = Boolean(image || payload)

  return (
    <>
      <div className="qris-action-box">
        <div className="qris-action-info">
          <span className="luxe-label">Total Payment</span>
          <strong className="luxe-value-strong text-primary">{amount}</strong>
          {description && <p className="text-muted small mb-0">{description}</p>}
        </div>
        <button className="btn btn-primary" type="button" disabled={!canOpen} onClick={() => setOpen(true)}>
          <i className="bi bi-qr-code-scan me-2"></i>
          View QRIS
        </button>
      </div>
      {open && (
        <div className="modal-backdrop-custom" role="dialog" aria-modal="true">
          <div className="modal-panel">
            <div className="modal-heading">
              <div>
                <h5>Scan QRIS</h5>
                <p>Pay the exact amount shown below.</p>
              </div>
              <button className="modal-close" type="button" onClick={() => setOpen(false)} aria-label="Close">
                <i className="bi bi-x-lg"></i>
              </button>
            </div>
            <div className="deposit-created">
              <div className="luxe-alert warning mb-2 text-start">
                 <i className="bi bi-shield-check"></i>
                 <div>
                    <strong className="d-block mb-1">Verify Merchant Name</strong>
                    <span>Please ensure you are paying to <strong>ZA Labs Tech</strong>.</span>
                 </div>
              </div>
              <div className="qris-modal-code">
                <QRISVisual image={image} payload={payload} />
              </div>
              <div className="deposit-created-total">
                <span>Total Payment</span>
                <strong>{amount}</strong>
              </div>
              <button className="btn btn-dark w-100 mt-3" type="button" onClick={() => setOpen(false)}>Done</button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default QRISPaymentBox
