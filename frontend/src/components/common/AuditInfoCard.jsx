const AuditInfoCard = ({ item }) => (
  <div className="luxe-detail-card bg-light border-0 shadow-none mt-4">
    <div className="luxe-card-body p-3">
      <div className="d-flex justify-content-between align-items-center">
        <div className="luxe-item">
          <span className="luxe-label" style={{ fontSize: '10px' }}>Created Record</span>
          <span className="luxe-value" style={{ fontSize: '13px' }}>
             <i className="bi bi-calendar-check me-1"></i>
             {item?.created_at ? new Date(item.created_at).toLocaleString('id-ID') : '-'}
          </span>
        </div>
        <div className="luxe-item text-end">
          <span className="luxe-label" style={{ fontSize: '10px' }}>Last System Update</span>
          <span className="luxe-value" style={{ fontSize: '13px' }}>
             {item?.updated_at ? new Date(item.updated_at).toLocaleString('id-ID') : '-'}
             <i className="bi bi-clock-history ms-1"></i>
          </span>
        </div>
      </div>
    </div>
  </div>
)

export default AuditInfoCard
