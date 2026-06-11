import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import roleService from '../../services/roleService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const RoleDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [role, setRole] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    roleService.getById(id)
      .then((response) => setRole(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load role')))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Role Detail</h1>
          <p>Access control levels and permissions</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/roles" />
          {hasPermission('roles', 'update') && (
            <Link to={`/roles/${id}/edit`} className="btn btn-primary d-flex align-items-center gap-2">
              <i className="bi bi-pencil-square"></i> Edit Role
            </Link>
          )}
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
            <i className="bi bi-shield-lock"></i> Access Control
          </div>
          <h2 className="luxe-hero-title">{role?.display_name || role?.name}</h2>
          <p className="luxe-hero-subtitle">{role?.description || 'No description provided'}</p>
        </div>
        <div className="luxe-hero-badge">
           <span className={`status-badge ${role?.is_system ? 'info' : 'success'}`}>
              {role?.is_system ? 'System Role' : 'Custom Role'}
           </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-info-circle"></i> Role Profile</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Unique Name</span>
                <span className="luxe-value-code">{role?.name || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Display Name</span>
                <span className="luxe-value">{role?.display_name || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Permission Count</span>
                <span className="luxe-value-strong">{role?.permission_ids?.length || 0} assigned</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Assigned Menus</span>
                <span className="luxe-value-strong">{role?.menu_ids?.length || 0} visible</span>
              </div>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-gear"></i> System Info</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
               <div className="luxe-item">
                  <span className="luxe-label">Immutable</span>
                  <span className="luxe-value">{role?.is_system ? 'Yes (System protected)' : 'No'}</span>
               </div>
               <div className="luxe-item">
                  <span className="luxe-label">Internal ID</span>
                  <span className="luxe-value-code">{role?.id}</span>
               </div>
            </div>
            
            {role?.is_system && (
              <div className="auth-alert mt-4">
                 <i className="bi bi-exclamation-triangle me-2"></i>
                 System roles are essential for application stability and may have editing restrictions.
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  )
}

export default RoleDetail
