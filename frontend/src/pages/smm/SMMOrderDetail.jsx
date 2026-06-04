import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage } from '../../services/api'

const SMMOrderDetail = () => {
  const { id } = useParams()
  const [order, setOrder] = useState(null)
  const [logs, setLogs] = useState([])

  const load = useCallback(async () => {
    const [orderRes, logsRes] = await Promise.allSettled([
      smmService.getOrder(id),
      smmService.getOrderStatusLogs(id),
    ])
    if (orderRes.status === 'fulfilled') setOrder(orderRes.value.data.data)
    if (logsRes.status === 'fulfilled') setLogs(logsRes.value.data.data || [])
  }, [id])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load order')))
  }, [load])

  const refresh = async () => {
    try {
      await smmService.refreshOrderStatus(id)
      toast.success('Status refreshed')
      await load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Refresh failed'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>SMM Order Detail</h1><p>{order?.ref_id || id}</p></div>
        <button className="btn btn-primary" onClick={refresh}>Refresh status</button>
      </div>
      <div className="content-grid two">
        <section className="panel">
          <div className="detail-grid">
            <span>Status</span><strong>{order?.status || '-'}</strong>
            <span>Target</span><strong>{order?.target || '-'}</strong>
            <span>Quantity</span><strong>{order?.quantity || '-'}</strong>
            <span>Service</span><strong>{order?.service?.name || order?.service_name || order?.provider_service_id || '-'}</strong>
            <span>Provider order</span><strong>{order?.provider_order_id || '-'}</strong>
            <span>Created</span><strong>{order?.created_at ? new Date(order.created_at).toLocaleString('id-ID') : '-'}</strong>
          </div>
        </section>
        <section className="panel">
          <div className="panel-heading"><h2>Status Logs</h2></div>
          {logs.map((log) => (
            <div className="money-row" key={log.id || `${log.new_status}-${log.created_at}`}>
              <span>{log.old_status ? `${log.old_status} -> ${log.new_status}` : log.new_status}</span>
              <strong>{log.created_at ? new Date(log.created_at).toLocaleString('id-ID') : '-'}</strong>
            </div>
          ))}
          {!logs.length && <p className="text-muted mb-0">No logs found</p>}
        </section>
      </div>
    </div>
  )
}

export default SMMOrderDetail
