const AuditInfoCard = ({ item }) => (
  <div className="info-card">
    <div className="text-muted small">Audit</div>
    <div>Created: {item?.created_at ? new Date(item.created_at).toLocaleString('id-ID') : '-'}</div>
    <div>Updated: {item?.updated_at ? new Date(item.updated_at).toLocaleString('id-ID') : '-'}</div>
  </div>
)

export default AuditInfoCard
