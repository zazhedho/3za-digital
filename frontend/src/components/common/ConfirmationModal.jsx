const ConfirmationModal = ({
  show,
  title,
  message,
  onConfirm,
  onCancel,
  confirmLabel = 'Confirm',
  confirmClassName = 'btn-primary',
  loading = false,
  confirmDisabled = false,
  children,
}) => {
  if (!show) return null

  return (
    <div className="modal-backdrop-custom" role="dialog" aria-modal="true">
      <div className="modal-panel confirmation-modal">
        <div className="modal-heading">
          <div>
            <h5>{title}</h5>
            <p>{message}</p>
          </div>
          <button type="button" className="modal-close" aria-label="Close" onClick={onCancel} disabled={loading}>
            <i className="bi bi-x-lg"></i>
          </button>
        </div>
        {children && <div className="confirmation-modal-body">{children}</div>}
        <div className="toolbar-actions justify-content-end">
          <button type="button" className="btn btn-outline-dark" onClick={onCancel} disabled={loading}>Cancel</button>
          <button type="button" className={`btn ${confirmClassName}`} onClick={onConfirm} disabled={loading || confirmDisabled}>
            {loading ? 'Processing...' : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}

export default ConfirmationModal
