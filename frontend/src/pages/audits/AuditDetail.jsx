import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import auditService from '../../services/auditService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const statusClass = (status) => {
  if (status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  return 'warning'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('id-ID')
}

const AuditDetail = () => {
  const { id } = useParams()
  const [audit, setAudit] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    auditService.getById(id)
      .then((response) => setAudit(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load audit trail')))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Audit Detail</h1>
          <p>System activity tracking and forensic logs</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/audits" />
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
            <i className="bi bi-shield-shaded"></i> {audit?.resource_label || audit?.resource} Log
          </div>
          <h2 className="luxe-hero-title">{audit?.summary || audit?.action_label || audit?.action}</h2>
          <p className="luxe-hero-subtitle">{audit?.message || 'No additional message recorded'}</p>
        </div>
        <div className="luxe-hero-badge">
          <span className={`status-badge ${statusClass(audit?.status)}`}>
            {audit?.status_label || audit?.status}
          </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-info-circle"></i> Event Context</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Action</span>
                <span className="luxe-value-strong text-capitalize">{audit?.action_label || audit?.action || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Resource Type</span>
                <span className="luxe-value text-capitalize">{audit?.resource_label || audit?.resource || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Resource ID</span>
                <span className="luxe-value-code">{audit?.resource_id || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Timestamp</span>
                <span className="luxe-value">{formatDate(audit?.occurred_at)}</span>
              </div>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-person-circle"></i> Actor & Origin</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Actor User ID</span>
                <span className="luxe-value-code">{audit?.actor?.user_id || 'System'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Actor Role</span>
                <span className="luxe-value text-capitalize">{audit?.actor?.role || 'System'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">IP Address</span>
                <span className="luxe-value">{audit?.ip_address || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Request ID</span>
                <span className="luxe-value-code">{audit?.request_id || '-'}</span>
              </div>
            </div>
            <div className="mt-4 pt-3 border-top">
               <span className="luxe-label mb-2 d-block">User Agent</span>
               <small className="text-muted text-break">{audit?.user_agent || '-'}</small>
            </div>
          </div>
        </section>
      </div>

      {(audit?.before_data || audit?.after_data) && (
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-braces"></i> Data Payload</h3>
          </div>
          <div className="luxe-card-body">
             <div className="content-grid two">
                <div>
                   <span className="luxe-label mb-2 d-block">Before Data</span>
                   <pre className="luxe-value-code" style={{ maxHeight: '300px', overflow: 'auto' }}>
                      {audit?.before_data ? JSON.stringify(audit.before_data, null, 2) : 'No previous data'}
                   </pre>
                </div>
                <div>
                   <span className="luxe-label mb-2 d-block">After Data</span>
                   <pre className="luxe-value-code" style={{ maxHeight: '300px', overflow: 'auto' }}>
                      {audit?.after_data ? JSON.stringify(audit.after_data, null, 2) : 'No changed data'}
                   </pre>
                </div>
             </div>
          </div>
        </section>
      )}

      {audit?.error_message && (
        <div className="auth-alert bg-danger text-white border-0">
           <strong>System Error:</strong>
           <p className="mb-0">{audit.error_message}</p>
        </div>
      )}
    </div>
  )
}

export default AuditDetail
