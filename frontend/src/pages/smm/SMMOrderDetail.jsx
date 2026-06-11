import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'
import { useAuth } from '../../contexts/AuthContext'

const formatDate = (value) => (value ? new Date(value).toLocaleString('id-ID') : '-')
const formatNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? new Intl.NumberFormat('id-ID').format(number) : '-'
}
const formatMoney = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(number) : '-'
}
const statusClass = (status = '') => {
  if (status === 'completed') return 'success'
  if (['pending', 'processing', 'partial'].includes(status)) return 'warning'
  if (['failed', 'cancelled'].includes(status)) return 'danger'
  return ''
}
const prettyJSON = (value) => {
  if (!value) return ''
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2)
    } catch {
      return value
    }
  }
  return JSON.stringify(value, null, 2)
}
const parseJSONValue = (value) => {
  if (!value) return {}
  if (typeof value === 'string') {
    try {
      return JSON.parse(value)
    } catch {
      return {}
    }
  }
  return value
}
const hasJSON = (value) => Boolean(prettyJSON(value).trim() && prettyJSON(value).trim() !== '{}')
const targetHref = (value) => {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.href : ''
  } catch {
    return ''
  }
}

const DetailItem = ({ label, children, code = false, strong = false }) => (
  <div className="luxe-item">
    <span className="luxe-label">{label}</span>
    <span className={code ? 'luxe-value-code' : strong ? 'luxe-value-strong' : 'luxe-value'}>{children ?? '-'}</span>
  </div>
)

const JSONCard = ({ title, value }) => {
  const rendered = prettyJSON(value)
  if (!rendered.trim() || rendered.trim() === '{}') return null
  return (
    <section className="audit-json-card">
      <h2>{title}</h2>
      <pre>{rendered}</pre>
    </section>
  )
}

const SMMOrderDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [order, setOrder] = useState(null)
  const [logs, setLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const canViewProviderData = hasPermission('smm_orders', 'list_all')

  const fetchData = useCallback(async () => {
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
  }, [id])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const refresh = async () => {
    setRefreshing(true)
    try {
      await smmService.refreshOrderStatus(id)
      await fetchData()
      toast.success('Order status updated')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Refresh failed'))
    } finally {
      setRefreshing(false)
    }
  }

  if (loading) return <div className="loading-fade">Loading...</div>

  const providerResponse = parseJSONValue(order?.provider_response)
  const providerData = providerResponse?.data || {}
  const serviceName = providerData.service_name || order?.service_name || order?.provider_service_id || '-'
  const orderNumber = providerData.order_number || order?.customer_no || ''
  const statusDescription = providerData.status_description || providerData.status_label || providerData.transaction_status || ''
  const targetUrl = targetHref(order?.target)

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
            <i className="bi bi-cart-check"></i> {order?.product_type || 'SMM'} Order
          </div>
          <h2 className="luxe-hero-title">{order?.ref_id || 'Order Detail'}</h2>
          <p className="luxe-hero-subtitle">{serviceName}</p>
        </div>
        <div className="luxe-hero-badge">
          <span className={`status-badge ${statusClass(order?.status)}`}>
            {order?.status || 'Unknown'}
          </span>
        </div>
      </div>

      <div className="content-grid two">
        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-hash"></i> Order Identity</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <DetailItem label="Reference" code>{order?.ref_id}</DetailItem>
              <DetailItem label="Order ID" code>{order?.id}</DetailItem>
              <DetailItem label="External Reference" code>{orderNumber || 'Not available'}</DetailItem>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-stars"></i> Service</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <DetailItem label="Service Name" strong>{serviceName}</DetailItem>
              {canViewProviderData && <DetailItem label="Service Code" code>{order?.provider_service_id}</DetailItem>}
              {canViewProviderData && <DetailItem label="Service ID" code>{order?.service_id || '-'}</DetailItem>}
              <DetailItem label="Category">{order?.service_category || '-'}</DetailItem>
              <DetailItem label="Platform">{order?.service_platform || '-'}</DetailItem>
              {canViewProviderData && <DetailItem label="Internal Provider">{order?.provider}</DetailItem>}
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-send-check"></i> Delivery</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <DetailItem label="Target" strong>
                {targetUrl ? <a href={targetUrl} target="_blank" rel="noreferrer">{order?.target}</a> : order?.target}
              </DetailItem>
              <DetailItem label="Quantity">{formatNumber(order?.quantity)}</DetailItem>
              <DetailItem label="Delivered">{formatNumber(providerData.delivered)}</DetailItem>
              <DetailItem label="Status Detail">{statusDescription || '-'}</DetailItem>
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-cash-coin"></i> Pricing</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <DetailItem label="User Paid" strong>{formatMoney(order?.amount)}</DetailItem>
              {canViewProviderData && <DetailItem label="Internal Cost">{formatMoney(order?.provider_charge)}</DetailItem>}
              {canViewProviderData && <DetailItem label="Raw Cost">{formatMoney(providerData.price)}</DetailItem>}
              {canViewProviderData && <DetailItem label="Profit">{formatMoney(order?.profit)}</DetailItem>}
              {canViewProviderData && <DetailItem label="Internal Balance After">{formatMoney(providerData.balance)}</DetailItem>}
            </div>
          </div>
        </section>

        <section className="luxe-detail-card">
          <div className="luxe-card-header">
            <h3><i className="bi bi-calendar3"></i> Timestamps</h3>
          </div>
          <div className="luxe-card-body">
            <div className="luxe-grid">
              <DetailItem label="Created At">{formatDate(order?.created_at)}</DetailItem>
              <DetailItem label="Updated At">{formatDate(order?.updated_at)}</DetailItem>
              {canViewProviderData && <DetailItem label="Created By" code>{order?.created_by || '-'}</DetailItem>}
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
                          <span className={`status-badge status-badge-sm ${statusClass(log.new_status)}`}>{log.new_status}</span>
                          <small>{formatDate(log.created_at)}</small>
                        </div>
                        {log.old_status && (
                          <p className="timeline-note">Status changed from <strong>{log.old_status}</strong></p>
                        )}
                        {log.provider_status && (
                          <p className="timeline-note">External status: <strong>{log.provider_status}</strong></p>
                        )}
                        {canViewProviderData && hasJSON(log.provider_response) && (
                          <pre className="timeline-json">{prettyJSON(log.provider_response)}</pre>
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

      {canViewProviderData && (hasJSON(order?.metadata) || hasJSON(order?.provider_response)) && (
        <div className="audit-json-grid order-json-grid">
          <JSONCard title="Metadata" value={order?.metadata} />
          <JSONCard title="Raw Response" value={order?.provider_response} />
        </div>
      )}
    </div>
  )
}

export default SMMOrderDetail
