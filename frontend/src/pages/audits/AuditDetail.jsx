import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import auditService from '../../services/auditService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const formatDate = (value) => {
  if (!value) return '-'
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

const statusClass = (status) => {
  if (status === 'success') return 'status-paid'
  if (status === 'failed') return 'status-failed'
  if (status === 'pending') return 'status-pending'
  return 'status-default'
}

const formatPayload = (value) => {
  if (!value) return 'No data'
  if (typeof value === 'string') return value
  return JSON.stringify(value, null, 2)
}

const AuditPayload = ({ title, value }) => (
  <div className="audit-json-card">
    <h2>{title}</h2>
    <pre>{formatPayload(value)}</pre>
  </div>
)

const AuditDetail = () => {
  const { id } = useParams()
  const [audit, setAudit] = useState(null)

  useEffect(() => {
    auditService.getById(id)
      .then((response) => setAudit(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load audit trail')))
  }, [id])

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Audit Detail</h1>
          <p>{audit?.summary || 'Audit trail record'}</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/audits" />
        </div>
      </div>

      <section className="panel">
        <div className="detail-grid detail-grid-compact">
          <span>Status</span>
          <div className="detail-value">
            <span className={`status-badge status-badge-detail ${statusClass(audit?.status)}`}>
              {audit?.status_label || audit?.status || '-'}
            </span>
          </div>
          <span>Action</span><strong>{audit?.action_label || audit?.action || '-'}</strong>
          <span>Resource</span><strong>{audit?.resource_label || audit?.resource || '-'}</strong>
          <span>Resource ID</span><strong>{audit?.resource_id || '-'}</strong>
          <span>Actor Role</span><strong>{audit?.actor?.role || '-'}</strong>
          <span>Actor User</span><strong>{audit?.actor?.user_id || 'System'}</strong>
          <span>Request ID</span><strong>{audit?.request_id || '-'}</strong>
          <span>IP Address</span><strong>{audit?.ip_address || '-'}</strong>
          <span>User Agent</span><strong>{audit?.user_agent || '-'}</strong>
          <span>Occurred</span><strong>{formatDate(audit?.occurred_at)}</strong>
          <span>Message</span><strong>{audit?.message || '-'}</strong>
          <span>Error</span><strong>{audit?.error_message || '-'}</strong>
        </div>
      </section>

      <section className="audit-json-grid">
        <AuditPayload title="Before Data" value={audit?.before_data} />
        <AuditPayload title="After Data" value={audit?.after_data} />
        <AuditPayload title="Metadata" value={audit?.metadata} />
      </section>
    </div>
  )
}

export default AuditDetail
