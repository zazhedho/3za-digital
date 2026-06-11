import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import appConfigService from '../../services/appConfigService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const ConfigDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [config, setConfig] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    appConfigService.getById(id)
      .then((response) => setConfig(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load config')))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Config Detail</h1>
          <p>System runtime configuration parameters</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/configs" />
          {hasPermission('configs', 'update') && (
            <Link to={`/configs/${id}/edit`} className="btn btn-primary d-flex align-items-center gap-2">
              <i className="bi bi-pencil-square"></i> Edit Config
            </Link>
          )}
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
            <i className="bi bi-gear-fill"></i> {config?.category || 'System'} Config
          </div>
          <h2 className="luxe-hero-title">{config?.display_name || config?.config_key}</h2>
          <p className="luxe-hero-subtitle">{config?.description || 'Operational parameter for the application'}</p>
        </div>
        <div className="luxe-hero-badge">
           <span className={`status-badge ${config?.is_active ? 'success' : 'danger'}`}>
              {config?.is_active ? 'Active' : 'Disabled'}
           </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-database"></i> Parameter Data</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Config Key</span>
                <span className="luxe-value-code">{config?.config_key || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Current Value</span>
                <span className="luxe-value-strong">{config?.value || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Category</span>
                <span className="luxe-value text-capitalize">{config?.category || '-'}</span>
              </div>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-clock-history"></i> Record Info</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
               <div className="luxe-item">
                  <span className="luxe-label">Created At</span>
                  <span className="luxe-value">{new Date(config?.created_at).toLocaleString('id-ID')}</span>
               </div>
               <div className="luxe-item">
                  <span className="luxe-label">Last Updated</span>
                  <span className="luxe-value">{config?.updated_at ? new Date(config.updated_at).toLocaleString('id-ID') : '-'}</span>
               </div>
               <div className="luxe-item">
                  <span className="luxe-label">Internal ID</span>
                  <span className="luxe-value-code">{config?.id}</span>
               </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}

export default ConfigDetail
