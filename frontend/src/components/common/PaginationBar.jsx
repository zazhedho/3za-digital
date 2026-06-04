const PaginationBar = ({ pagination, onPageChange, loading }) => {
  const total = pagination?.total || 0
  const page = pagination?.page || 1
  const totalPages = pagination?.totalPages || 1
  const limit = pagination?.limit || 0
  const start = total === 0 ? 0 : ((page - 1) * limit) + 1
  const end = Math.min(page * limit, total)

  return (
    <div className="pagination-bar">
      <div className="pagination-summary">
        {total ? `${start}-${end} of ${total}` : '0 results'}
      </div>
      <div className="pagination-actions">
        <button
          className="btn btn-outline-dark btn-sm"
          type="button"
          disabled={loading || page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          <i className="bi bi-chevron-left"></i>
        </button>
        <span>Page {page} / {totalPages}</span>
        <button
          className="btn btn-outline-dark btn-sm"
          type="button"
          disabled={loading || page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          <i className="bi bi-chevron-right"></i>
        </button>
      </div>
    </div>
  )
}

export default PaginationBar
