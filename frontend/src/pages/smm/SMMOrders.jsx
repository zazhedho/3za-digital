import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'

const SMMOrders = () => {
  const [params] = useSearchParams()
  const [rows, setRows] = useState([])
  const [search, setSearch] = useState(params.get('search') || '')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 50 })
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setSearch(params.get('search') || '')
    setPage(1)
  }, [params])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await smmService.getOrders({
        search,
        page,
        limit: 50,
        'filters[status]': status || undefined,
      })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load orders'))
    } finally {
      setLoading(false)
    }
  }, [page, search, status])

  useEffect(() => {
    load()
  }, [load])

  const refresh = async (id) => {
    try {
      await smmService.refreshOrderStatus(id)
      toast.success('Status refreshed')
      load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Refresh failed'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>SMM Orders</h1>
          <p>Order status, target, quantity, and provider references.</p>
        </div>
        <Link to="/smm/orders/new" className="btn btn-primary"><i className="bi bi-plus-lg me-2"></i>New order</Link>
      </div>

      <div className="filter-pill compact status-filter">
        <i className="bi bi-funnel"></i>
        <input value={search} onChange={(event) => { setSearch(event.target.value); setPage(1) }} placeholder="Search ref, target, or service" />
        <select value={status} onChange={(event) => { setStatus(event.target.value); setPage(1) }}>
          <option value="">All status</option>
          <option value="pending">Pending</option>
          <option value="processing">Processing</option>
          <option value="completed">Completed</option>
          <option value="partial">Partial</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <button className="btn btn-dark" onClick={load}>Apply</button>
      </div>

      <section className="table-panel">
        <table className="table align-middle">
          <thead>
            <tr>
              <th>Reference</th>
              <th>Service</th>
              <th>Target</th>
              <th>Quantity</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <strong>{row.ref_id || row.id}</strong>
                  <div className="text-muted small">{row.provider_order_id || '-'}</div>
                </td>
                <td>{row.service?.name || row.service_name || row.provider_service_id || '-'}</td>
                <td className="text-truncate" style={{ maxWidth: 260 }}>{row.target}</td>
                <td>{new Intl.NumberFormat('id-ID').format(row.quantity || 0)}</td>
                <td><span className="badge rounded-pill text-bg-light text-capitalize">{row.status || '-'}</span></td>
                <td className="text-end">
                  <Link className="btn btn-sm btn-outline-dark me-2" to={`/smm/orders/${row.id}`}>Detail</Link>
                  <button className="btn btn-sm btn-outline-dark" onClick={() => refresh(row.id)}>Refresh</button>
                </td>
              </tr>
            ))}
            {!rows.length && (
              <tr><td colSpan="6" className="text-center text-muted py-5">{loading ? 'Loading...' : 'No orders found'}</td></tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default SMMOrders
