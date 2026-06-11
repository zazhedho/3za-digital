import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'
import TableActionMenu from '../../components/common/TableActionMenu'

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
      toast.success('Status refreshed from provider')
      load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Refresh failed'))
    } finally {
      setRefreshingId('')
    }
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>SMM Orders</h1>
          <p>Track your digital service delivery and status updates.</p>
        </div>
        <Link to="/smm/orders/new" className="btn btn-primary d-flex align-items-center gap-2">
           <i className="bi bi-plus-circle-fill"></i> New Order
        </Link>
      </div>

      <div className="toolbar-actions mb-4">
        <form className="filter-pill filter-only status-filter" onSubmit={submitSearch}>
          <i className="bi bi-funnel"></i>
          <input 
            value={searchInput} 
            onChange={(event) => setSearchInput(event.target.value)} 
            placeholder="Search reference or target..." 
          />
          <select value={statusInput} onChange={(event) => setStatusInput(event.target.value)}>
            <option value="">All Status</option>
            <option value="pending">Pending</option>
            <option value="processing">Processing</option>
            <option value="completed">Completed</option>
            <option value="partial">Partial</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <button className="btn btn-dark" type="submit" disabled={loading}>Filter</button>
          <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
            <i className="bi bi-x-lg me-2"></i>Reset
          </button>
        </form>
      </div>

      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Order ID</th>
              <th>Service Item</th>
              <th>Target Link</th>
              <th className="text-end">Quantity</th>
              <th>Status</th>
              <th className="text-end">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <span className="table-main">
                    <strong>{row.ref_id || row.id}</strong>
                    <span className="table-subtext">{row.customer_no ? `ID: ${row.customer_no}` : 'Waiting for ID...'}</span>
                  </span>
                </td>
                <td>
                   <div className="table-main">
                      <span className="table-subtext d-block">{row.service?.name || row.service_name || row.provider_service_id || '-'}</span>
                      <small className="status-badge status-badge-sm info text-capitalize">{row.product_type || 'SMM'}</small>
                   </div>
                </td>
                <td><span className="table-long text-primary" title={row.target}>{row.target}</span></td>
                <td className="table-number"><strong>{new Intl.NumberFormat('id-ID').format(row.quantity || 0)}</strong></td>
                <td>
                   <span className={`status-badge ${row.status === 'completed' ? 'success' : row.status === 'pending' || row.status === 'processing' ? 'warning' : 'danger'} text-capitalize`}>
                      {row.status || 'unknown'}
                   </span>
                </td>
                <td className="text-end">
                   <TableActionMenu
                      label="Order actions"
                      items={[
                        { label: 'View Detail', to: `/smm/orders/${row.id}` },
                        { label: refreshingId === row.id ? 'Refreshing...' : 'Sync Status', onClick: () => refresh(row), disabled: refreshingId === row.id },
                      ]}
                   />
                </td>
              </tr>
            ))}
            {!rows.length && (
              <tr>
                <td colSpan="6" className="empty-cell py-5">
                   <div className="text-center">
                      <i className="bi bi-cart-x text-muted display-4"></i>
                      <p className="mt-3 text-muted">{loading ? 'Loading orders...' : 'No orders found.'}</p>
                   </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default SMMOrders
