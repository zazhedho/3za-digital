import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const SMMOrderDetail = () => {
  const { id } = useParams()
  const [order, setOrder] = useState(null)
  const [logs, setLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const fetchData = async () => {
    try {
      const [orderRes, logsRes] = await Promise.all([
        smmService.getOrder(id),
        smmService.getOrderStatusLogs(id)
      ])
      setOrder(orderRes.data.data)
      setLogs(logsRes.data.data || [])
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load order details'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [id])

  const refresh = async () => {
    setRefreshing(true)
    try {
      await smmService.refreshOrderStatus(id)
      await fetchData()
      toast.success('Status updated from provider')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Refresh failed'))
    } finally {
      setRefreshing(false)
    }
  }

  if (loading) return <div className="loading-fade">Loading...</div>

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>SMM Order Detail</h1>
          <p>Service delivery tracking and status logs</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/smm/orders" />
          <button className="btn btn-primary d-flex align-items-center gap-2" onClick={refresh} disabled={refreshing}>
            <i className={`bi ${refreshing ? 'bi-arrow-repeat spin' : 'bi-arrow-repeat'}`}></i>
            {refreshing ? 'Syncing...' : 'Sync Status'}
          </button>
        </div>
      </div>

      <div className="luxe-detail-hero">
        <div className="luxe-hero-content">
          <div className="luxe-hero-kicker">
            <i className="bi bi-cart-check"></i> {order?.service?.product_type || 'SMM'} Order
          </div>
          <h2 className="luxe-hero-title">{order?.ref_id || 'Order Detail'}</h2>
          <p className="luxe-hero-subtitle">{order?.service?.name || order?.service_name}</p>
        </div>
        <div className="luxe-hero-badge">
          <span className={`status-badge ${order?.status === 'completed' ? 'success' : order?.status === 'pending' || order?.status === 'processing' ? 'warning' : 'danger'}`}>
            {order?.status || 'Unknown'}
          </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-info-circle"></i> Service Details</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <div className="luxe-item">
                <span className="luxe-label">Target Link</span>
                <span className="luxe-value-strong">{order?.target || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Quantity</span>
                <span className="luxe-value">{order?.quantity?.toLocaleString('id-ID') || '-'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Provider Order ID</span>
                <span className="luxe-value-code">{order?.customer_no || 'Not assigned'}</span>
              </div>
              <div className="luxe-item">
                <span className="luxe-label">Created At</span>
                <span className="luxe-value">{order?.created_at ? new Date(order.created_at).toLocaleString('id-ID') : '-'}</span>
              </div>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-clock-history"></i> Status Logs</h3>
          </div>
          <div className="luxe-card-body">
             {logs.length > 0 ? (
               <div className="timeline-luxe">
                  {logs.map((log, index) => (
                    <div className="timeline-item" key={log.id || index}>
                      <div className="timeline-dot"></div>
                      <div className="timeline-content">
                        <div className="timeline-header">
                          <span className="status-badge status-badge-sm warning">{log.new_status}</span>
                          <small>{new Date(log.created_at).toLocaleString('id-ID')}</small>
                        </div>
                        {log.old_status && (
                          <p className="timeline-note">Status changed from <strong>{log.old_status}</strong></p>
                        )}
                      </div>
                    </div>
                  ))}
               </div>
             ) : (
               <div className="text-center py-4">
                  <i className="bi bi-inbox text-muted display-4"></i>
                  <p className="text-muted mt-2">No status logs recorded yet.</p>
               </div>
             )}
          </div>
        </section>
      </div>
    </div>
  )
}

export default SMMOrderDetail
