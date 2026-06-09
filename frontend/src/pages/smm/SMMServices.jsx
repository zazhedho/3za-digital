import { Fragment, useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import smmService from '../../services/smmService'
import { getErrorMessage, getListPayload } from '../../services/api'
import PaginationBar from '../../components/common/PaginationBar'
import ConfirmationModal from '../../components/common/ConfirmationModal'
import { useAuth } from '../../contexts/AuthContext'

const formatMoney = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

const serviceGroupName = (row) => row.category || row.brand || 'Uncategorized'

const sortIconByDirection = (active, direction) => {
  if (!active) return 'bi-arrow-down-up'
  return direction === 'desc' ? 'bi-sort-down' : 'bi-sort-up'
}

const SMMServices = () => {
  const { hasPermission } = useAuth()
  const [params, setParams] = useSearchParams()
  const [rows, setRows] = useState([])
  const [searchInput, setSearchInput] = useState(params.get('search') || '')
  const [search, setSearch] = useState(params.get('search') || '')
  const [platformInput, setPlatformInput] = useState('')
  const [platform, setPlatform] = useState('')
  const [sortBy, setSortBy] = useState('category')
  const [sortDirection, setSortDirection] = useState('asc')
  const [page, setPage] = useState(1)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 50 })
  const [loading, setLoading] = useState(false)
  const [confirmSync, setConfirmSync] = useState(false)
  const [confirmLoading, setConfirmLoading] = useState(false)

  const groupedRows = useMemo(() => {
    const groups = []
    const groupIndex = new Map()

    rows.forEach((row) => {
      const name = serviceGroupName(row)
      if (!groupIndex.has(name)) {
        groupIndex.set(name, groups.length)
        groups.push({ name, rows: [] })
      }
      groups[groupIndex.get(name)].rows.push(row)
    })

    return groups
  }, [rows])

  useEffect(() => {
    const nextSearch = params.get('search') || ''
    setSearchInput(nextSearch)
    setSearch(nextSearch)
    setSortBy(params.get('order_by') || 'category')
    setSortDirection((params.get('order_direction') || 'asc').toLowerCase() === 'desc' ? 'desc' : 'asc')
    setPage(1)
  }, [params])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const response = await smmService.getServices({
        search,
        page,
        limit: 50,
        order_by: sortBy,
        order_direction: sortDirection,
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
  }, [page, platform, search, sortBy, sortDirection])

  useEffect(() => {
    load()
  }, [load])

  const submitSearch = (event) => {
    event.preventDefault()
    setSearch(searchInput.trim())
    setPlatform(platformInput)
    setPage(1)
  }

  const resetSearch = () => {
    setSearchInput('')
    setSearch('')
    setPlatformInput('')
    setPlatform('')
    setSortBy('category')
    setSortDirection('asc')
    setPage(1)
    setParams({})
  }

  const changeSort = (field) => {
    setPage(1)
    if (sortBy === field) {
      setSortDirection((direction) => (direction === 'asc' ? 'desc' : 'asc'))
      return
    }
    setSortBy(field)
    setSortDirection('asc')
  }

  const sortButton = (field, label) => (
    <button
      type="button"
      className={`table-sort-trigger${sortBy === field ? ' active' : ''}`}
      onClick={() => changeSort(field)}
      aria-label={`Sort by ${label}`}
      aria-pressed={sortBy === field}
    >
      <span>{label}</span>
      <i className={`bi ${sortIconByDirection(sortBy === field, sortDirection)}`}></i>
    </button>
  )

  const sync = async () => {
    setConfirmLoading(true)
    try {
      await smmService.syncServices({ platform: platformInput })
      toast.success('Services synced')
      setConfirmSync(false)
      load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Sync failed'))
    } finally {
      setConfirmLoading(false)
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
          <button className="btn btn-primary" onClick={() => setConfirmSync(true)}><i className="bi bi-cloud-arrow-down me-2"></i>Sync</button>
        )}
      </div>

      <form className="filter-pill" onSubmit={submitSearch}>
        <i className="bi bi-search"></i>
        <input value={searchInput} onChange={(event) => setSearchInput(event.target.value)} placeholder="Search service name" />
        <select value={platformInput} onChange={(event) => setPlatformInput(event.target.value)}>
          <option value="">All platforms</option>
          <option value="instagram">Instagram</option>
          <option value="tiktok">TikTok</option>
          <option value="youtube">YouTube</option>
          <option value="facebook">Facebook</option>
        </select>
        <button className="btn btn-dark" type="submit" disabled={loading}>Search</button>
        <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
          <i className="bi bi-x-lg me-2"></i>Reset
        </button>
      </form>

      <section className="table-panel">
        <table className="table app-table table-wide align-middle">
          <thead>
            <tr>
              <th>{sortButton('provider_service_id', 'ID')}</th>
              <th>{sortButton('name', 'Service Name')}</th>
              <th>{sortButton('platform', 'Platform')}</th>
              <th>{sortButton('min_quantity', 'Min/Max')}</th>
              <th>{sortButton('price', 'Price / 1k')}</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {groupedRows.map((group) => (
              <Fragment key={group.name}>
                <tr className="table-group-row">
                  <td colSpan="6">
                    <div className="table-group-head">
                      <div className="table-group-meta">
                        <span className="table-group-title">{group.name}</span>
                        <span className="table-group-count">{group.rows.length} service{group.rows.length === 1 ? '' : 's'}</span>
                        {platform && <span className="table-group-badge text-capitalize">{platform}</span>}
                      </div>
                      {sortButton('category', 'Category')}
                    </div>
                  </td>
                </tr>
                {group.rows.map((row) => (
                  <tr key={row.id}>
                    <td><span className="table-id">{row.provider_service_id}</span></td>
                    <td><span className="table-main"><strong>{row.name}</strong></span></td>
                    <td className="text-capitalize table-nowrap">{row.platform || '-'}</td>
                    <td className="table-number">{row.min_quantity || '-'} / {row.max_quantity || '-'}</td>
                    <td className="table-number">{formatMoney(row.price)}</td>
                    <td><span className={`badge ${row.is_active ? 'text-bg-success' : 'text-bg-secondary'}`}>{row.is_active ? 'Active' : 'Inactive'}</span></td>
                  </tr>
                ))}
              </Fragment>
            ))}
            {!rows.length && (
              <tr><td colSpan="6" className="empty-cell">{loading ? 'Loading...' : 'No services found'}</td></tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
      <ConfirmationModal
        show={confirmSync}
        title="Sync Services"
        message={`Sync SMM services${platformInput ? ` for ${platformInput}` : ''} from provider? Existing service data may be updated.`}
        confirmLabel="Sync"
        loading={confirmLoading}
        onCancel={() => setConfirmSync(false)}
        onConfirm={sync}
      />
    </div>
  )
}

export default SMMServices
