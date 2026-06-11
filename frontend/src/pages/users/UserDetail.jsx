import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import userService from '../../services/userService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const UserDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    userService.getById(id)
      .then((response) => setUser(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load user')))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>User Detail</h1>
          <p>Member profile and account status</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/users" />
          {hasPermission('users', 'update') && (
            <Link to={`/users/${id}/edit`} className="btn btn-primary d-flex align-items-center gap-2">
              <i className="bi bi-pencil-square"></i> Edit Profile
            </Link>
          )}
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content d-flex align-items-center gap-4">
          <div className="user-avatar-luxe">
            {user?.avatar_url ? (
              <img src={user.avatar_url} alt={user.name} />
            ) : (
              <div className="avatar-placeholder">{user?.name?.charAt(0) || '?'}</div>
            )}
          </div>
          <div>
            <div className="luxe-hero-kicker">
              <i className="bi bi-person-badge"></i> {user?.role || 'Member'}
            </div>
            <h2 className="luxe-hero-title">{user?.name || 'Unknown User'}</h2>
            <p className="luxe-hero-subtitle">{user?.email}</p>
          </div>
        </div>
        <div className="luxe-hero-badge">
           <span className={`status-badge ${user?.is_active !== false ? 'success' : 'danger'}`}>
              {user?.is_active !== false ? 'Active Account' : 'Inactive'}
           </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-card-heading"></i> Account Information</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Full Name</span>
                <span className="luxe-value">{user?.name || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Email Address</span>
                <span className="luxe-value">{user?.email || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Phone Number</span>
                <span className="luxe-value">{user?.phone || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">User Role</span>
                <span className="luxe-value text-capitalize">{user?.role || '-'}</span>
              </div>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-shield-lock"></i> Security & Activity</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">User ID</span>
                <span className="luxe-value-code">{user?.id}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Joined Since</span>
                <span className="luxe-value">{user?.created_at ? new Date(user.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) : '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Last Activity</span>
                <span className="luxe-value">{user?.last_login_at ? new Date(user.last_login_at).toLocaleString('id-ID') : 'Never'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Account Security</span>
                <span className="luxe-value d-flex align-items-center gap-2">
                   <i className="bi bi-check-circle-fill text-success"></i> Verified
                </span>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}

export default UserDetail
