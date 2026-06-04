const ConfirmationModal = ({ show, title, message, onConfirm, onCancel, confirmLabel = 'Confirm' }) => {
  if (!show) return null

  return (
    <div className="modal-backdrop-custom">
      <div className="modal-panel">
        <h5>{title}</h5>
        <p>{message}</p>
        <div className="d-flex gap-2 justify-content-end">
          <button type="button" className="btn btn-outline-secondary" onClick={onCancel}>Cancel</button>
          <button type="button" className="btn btn-primary" onClick={onConfirm}>{confirmLabel}</button>
        </div>
      </div>
    </div>
  )
}

export default ConfirmationModal
