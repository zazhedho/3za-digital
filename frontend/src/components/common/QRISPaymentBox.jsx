import { useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'

const QRISVisual = ({ image, payload, size = 180 }) => {
  if (image) return <img src={image} alt="QRIS payment code" />
  if (payload) return <QRCodeSVG value={payload} size={size} level="M" includeMargin />
  return <div className="qris-placeholder">QRIS is not available yet</div>
}

const QRISPaymentBox = ({ amount, description, image, payload }) => {
  const [open, setOpen] = useState(false)
  const canOpen = Boolean(image || payload)

  return (
    <>
      <div className="qris-summary">
        <div>
          <span>Total Payment</span>
          <strong>{amount}</strong>
          {description && <p>{description}</p>}
        </div>
        <button className="btn btn-primary" type="button" disabled={!canOpen} onClick={() => setOpen(true)}>
          View QRIS
        </button>
      </div>
      {open && (
        <div className="modal-backdrop-custom" role="dialog" aria-modal="true">
          <div className="modal-panel qris-modal">
            <div className="modal-heading">
              <div>
                <h5>Scan QRIS</h5>
                <p>Pay the exact amount shown.</p>
              </div>
              <button className="modal-close" type="button" onClick={() => setOpen(false)} aria-label="Close">
                <i className="bi bi-x-lg"></i>
              </button>
            </div>
            <div className="qris-modal-body">
              <div className="qris-modal-code">
                <QRISVisual image={image} payload={payload} size={260} />
              </div>
              <strong>{amount}</strong>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default QRISPaymentBox
