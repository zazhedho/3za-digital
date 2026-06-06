import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'

const SMMOrders = () => {
  const [params, setParams] = useSearchParams()
  const [rows, setRows] = useState([])
  const [searchInput, setSearchInput] = useState(params.get('search') || '')
  const [search, setSearch] = useState(params.get('search') || '')
  const [statusInput, setStatusInput] = useState('')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 50 })
  const [loading, setLoading] = useState(false)
  const [refreshingId, setRefreshingId] = useState('')

  useEffect(() => {
    const nextSearch = params.get('search') || ''
    setSearchInput(nextSearch)
    setSearch(nextSearch)
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

  const submitSearch = (event) => {
    event.preventDefault()
    setSearch(searchInput.trim())
    setStatus(statusInput)
    setPage(1)
  }

  const resetSearch = () => {
    setSearchInput('')
    setSearch('')
    setStatusInput('')
    setStatus('')
    setPage(1)
    setParams({})
  }

  const refresh = async (order) => {
    if (!order?.id) return
    setRefreshingId(order.id)
    try {
      await smmService.refreshOrderStatus(order.id)
      toast.success('Status refreshed')
      load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Refresh failed'))
    } finally {
      setRefreshingId('')
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

      <form className="filter-pill compact status-filter" onSubmit={submitSearch}>
        <i className="bi bi-funnel"></i>
        <input value={searchInput} onChange={(event) => setSearchInput(event.target.value)} placeholder="Search ref, target, or service" />
        <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
          <option value="">All status</option>
          <option value="pending">Pending</option>
          <option value="processing">Processing</option>
          <option value="completed">Completed</option>
          <option value="partial">Partial</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <button className="btn btn-dark" type="submit" disabled={loading}>Search</button>
        <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
          <i className="bi bi-x-lg me-2"></i>Reset
        </button>
      </form>

      <section className="table-panel">
        <table className="table app-table align-middle">
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
                  <span className="table-main">
                    <strong>{row.ref_id || row.id}</strong>
                    <span className="table-subtext">{row.provider_order_id || '-'}</span>
                  </span>
                </td>
                <td><span className="table-subtext">{row.service?.name || row.service_name || row.provider_service_id || '-'}</span></td>
                <td><span className="table-long">{row.target}</span></td>
                <td className="table-number">{new Intl.NumberFormat('id-ID').format(row.quantity || 0)}</td>
                <td><span className="badge rounded-pill text-bg-light text-capitalize">{row.status || '-'}</span></td>
                <td className="text-end">
                  <span className="table-actions">
                    <Link className="btn btn-sm btn-outline-dark" to={`/smm/orders/${row.id}`}>Detail</Link>
                    <button className="btn btn-sm btn-outline-dark" onClick={() => refresh(row)} disabled={refreshingId === row.id}>
                      {refreshingId === row.id ? 'Refreshing...' : 'Refresh'}
                    </button>
                  </span>
                </td>
              </tr>
            ))}
            {!rows.length && (
              <tr><td colSpan="6" className="empty-cell">{loading ? 'Loading...' : 'No orders found'}</td></tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default SMMOrders
