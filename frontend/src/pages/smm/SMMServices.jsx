import { useCallback, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const SMMServices = () => {
  const [params] = useSearchParams()
  const [rows, setRows] = useState([])
  const [search, setSearch] = useState(params.get('search') || '')
  const [platform, setPlatform] = useState('')
  const [loading, setLoading] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await smmService.getServices({
        search,
        limit: 50,
        'filters[platform]': platform || undefined,
        'filters[is_active]': 'true',
      })
      setRows(getListPayload(response).rows)
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load services'))
    } finally {
      setLoading(false)
    }
  }, [platform, search])

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
        <button className="btn btn-primary" onClick={sync}><i className="bi bi-cloud-arrow-down me-2"></i>Sync</button>
      </div>

      <div className="filter-pill">
        <i className="bi bi-search"></i>
        <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search service name" />
        <select value={platform} onChange={(event) => setPlatform(event.target.value)}>
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
              <th>Service</th>
              <th>Platform</th>
              <th>Category</th>
              <th>Min/Max</th>
              <th>Price</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <strong>{row.name}</strong>
                  <div className="text-muted small">{row.provider_service_id}</div>
                </td>
                <td>{row.platform || '-'}</td>
                <td>{row.category || row.brand || '-'}</td>
                <td>{row.min_quantity || '-'} / {row.max_quantity || '-'}</td>
                <td>{formatMoney(row.price)}</td>
                <td><span className={`badge ${row.is_active ? 'text-bg-success' : 'text-bg-secondary'}`}>{row.is_active ? 'Active' : 'Inactive'}</span></td>
              </tr>
            ))}
            {!rows.length && (
              <tr><td colSpan="6" className="text-center text-muted py-5">{loading ? 'Loading...' : 'No services found'}</td></tr>
            )}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default SMMServices
