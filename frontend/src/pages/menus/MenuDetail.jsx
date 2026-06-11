import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import menuService from '../../services/menuService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const MenuDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [menu, setMenu] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    menuService.getById(id)
      .then((response) => setMenu(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load menu')))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>Menu Detail</h1>
          <p>Sidebar navigation and visibility settings</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/menus" />
          {hasPermission('menus', 'update') && (
            <Link to={`/menus/${id}/edit`} className="btn btn-primary d-flex align-items-center gap-2">
              <i className="bi bi-pencil-square"></i> Edit Menu
            </Link>
          )}
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content d-flex align-items-center gap-4">
          <div className="user-avatar-luxe" style={{ background: '#f8fafc', color: 'var(--primary)' }}>
             <i className={`bi bi-${menu?.icon || 'list'} display-4`}></i>
          </div>
          <div>
            <div className="luxe-hero-kicker">
              <i className="bi bi-layout-sidebar"></i> Navigation Item
            </div>
            <h2 className="luxe-hero-title">{menu?.display_name || menu?.name}</h2>
            <p className="luxe-hero-subtitle">Route: <span className="luxe-value-code">{menu?.path || '-'}</span></p>
          </div>
        </div>
        <div className="luxe-hero-badge">
           <span className={`status-badge ${menu?.is_active ? 'success' : 'danger'}`}>
              {menu?.is_active ? 'Visible' : 'Hidden'}
           </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-info-circle"></i> Basic Properties</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Unique Name</span>
                <span className="luxe-value-code">{menu?.name || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Display Label</span>
                <span className="luxe-value">{menu?.display_name || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Sort Order</span>
                <span className="luxe-value-strong">{menu?.order_index ?? '0'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Icon Class</span>
                <span className="luxe-value"><i className={`bi bi-${menu?.icon} me-2`}></i> {menu?.icon || '-'}</span>
              </div>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-clock-history"></i> System Metadata</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
               <div className="luxe-item">
                  <span className="luxe-label">Created At</span>
                  <span className="luxe-value">{new Date(menu?.created_at).toLocaleString('id-ID')}</span>
               </div>
               <div className="luxe-item">
                  <span className="luxe-label">Internal ID</span>
                  <span className="luxe-value-code">{menu?.id}</span>
               </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}

export default MenuDetail
