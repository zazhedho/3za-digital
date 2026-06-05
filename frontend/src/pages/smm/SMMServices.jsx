import { useCallback, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'
import { useAuth } from '../../contexts/AuthContext'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const SMMServices = () => {
  const { hasPermission } = useAuth()
  const [params] = useSearchParams()
  const [rows, setRows] = useState([])
  const [search, setSearch] = useState(params.get('search') || '')
  const [platform, setPlatform] = useState('')
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
      const response = await smmService.getServices({
        search,
        page,
        limit: 50,
        'filters[platform]': platform || undefined,
        'filters[is_active]': 'true',
      })
      const payload = getListPayload(response)
      setRows(payload.rows)
      setPagination(payload)
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load services'))
    } finally {
      setLoading(false)
    }
  }, [page, platform, search])

  useEffect(() => {
    load()
  }, [load])

  const sync = async () => {
    try {
      await smmService.syncServices({ platform })
      toast.success('Services synced')
      load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Sync failed'))
    }
  }

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>SMM Services</h1>
          <p>Provider catalog for SMM orders.</p>
        </div>
        {hasPermission('smm_services', 'sync') && (
          <button className="btn btn-primary" onClick={sync}><i className="bi bi-cloud-arrow-down me-2"></i>Sync</button>
        )}
      </div>

      <div className="filter-pill">
        <i className="bi bi-search"></i>
        <input value={search} onChange={(event) => { setSearch(event.target.value); setPage(1) }} placeholder="Search service name" />
        <select value={platform} onChange={(event) => { setPlatform(event.target.value); setPage(1) }}>
          <option value="">All platforms</option>
          <option value="instagram">Instagram</option>
          <option value="tiktok">TikTok</option>
          <option value="youtube">YouTube</option>
          <option value="facebook">Facebook</option>
        </select>
        <button className="btn btn-dark" onClick={load}>Apply</button>
      </div>

      <section className="table-panel">
        <table className="table align-middle">
          <thead>
            <tr>
              <th>ID</th>
              <th>Service Name</th>
              <th>Platform</th>
              <th>Category</th>
              <th>Min/Max</th>
              <th>Price / 1k</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><span className="text-muted small">{row.provider_service_id}</span></td>
                <td><strong>{row.name}</strong></td>
                <td>{row.platform || '-'}</td>
                <td>{row.category || row.brand || '-'}</td>
                <td>{row.min_quantity || '-'} / {row.max_quantity || '-'}</td>
                <td>{formatMoney(row.price)}</td>
                <td><span className={`badge ${row.is_active ? 'text-bg-success' : 'text-bg-secondary'}`}>{row.is_active ? 'Active' : 'Inactive'}</span></td>
              </tr>
            ))}
            {!rows.length && (
              <tr><td colSpan="7" className="text-center text-muted py-5">{loading ? 'Loading...' : 'No services found'}</td></tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
    </div>
  )
}

export default SMMServices
