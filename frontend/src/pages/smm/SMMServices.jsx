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
  const [limit, setLimit] = useState(20)
  const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1, limit: 20 })
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
        limit,
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
  }, [page, limit, platform, search, sortBy, sortDirection])

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
    setLimit(20)
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
      toast.success('Provider services synced successfully')
      setConfirmSync(false)
      load()
    } catch (error) {
      toast.error(getErrorMessage(error, 'Sync failed'))
    } finally {
      setConfirmLoading(false)
    }
  }

  return (
    <div className="luxe-page-fade">
      <div className="page-toolbar">
        <div>
          <h1>SMM Services</h1>
          <p>Global service catalog and real-time pricing.</p>
        </div>
        {hasPermission('smm_services', 'sync') && (
          <button className="btn btn-primary d-flex align-items-center gap-2" onClick={() => setConfirmSync(true)}>
             <i className="bi bi-cloud-arrow-down-fill"></i> Sync Provider
          </button>
        )}
      </div>

      <div className="toolbar-actions list-filter-bar mb-4">
        <form className="filter-pill filter-only" onSubmit={submitSearch}>
          <i className="bi bi-search"></i>
          <input 
             value={searchInput} 
             onChange={(event) => setSearchInput(event.target.value)} 
             placeholder="Search service name..." 
             style={{ flex: 2 }}
          />
          <select value={platformInput} onChange={(event) => setPlatformInput(event.target.value)}>
            <option value="">All Platforms</option>
            <option value="instagram">Instagram</option>
            <option value="tiktok">TikTok</option>
            <option value="youtube">YouTube</option>
            <option value="facebook">Facebook</option>
          </select>
          <select value={limit} onChange={(event) => { setLimit(Number(event.target.value)); setPage(1); }}>
            <option value="10">Show 10</option>
            <option value="20">Show 20</option>
            <option value="30">Show 30</option>
            <option value="50">Show 50</option>
            <option value="100">Show 100</option>
          </select>
          <button className="btn btn-dark px-4" type="submit" disabled={loading}>Filter</button>
          <button className="btn btn-outline-dark" type="button" onClick={resetSearch} disabled={loading}>
            <i className="bi bi-x-lg me-2"></i>Reset
          </button>
        </form>
      </div>

      <section className="table-panel">
        <table className="table app-table table-wide align-middle">
          <thead>
            <tr>
              <th style={{ width: '80px' }}>{sortButton('provider_service_id', 'ID')}</th>
              <th>{sortButton('name', 'Service Name')}</th>
              <th>{sortButton('platform', 'Platform')}</th>
              <th className="text-end">Min / Max</th>
              <th className="text-end">{sortButton('price', 'Price / 1k')}</th>
              <th className="text-end">Status</th>
            </tr>
          </thead>
          <tbody>
            {groupedRows.map((group) => (
              <Fragment key={group.name}>
                <tr className="table-group-row">
                  <td colSpan="6">
                    <div className="table-group-head">
                      <div className="table-group-meta">
                        <i className="bi bi-folder2-open text-primary"></i>
                        <span className="table-group-title">{group.name}</span>
                        <span className="table-group-count">{group.rows.length} items</span>
                      </div>
                      <span className="status-badge status-badge-sm info text-capitalize">{group.rows[0]?.platform || 'multi'}</span>
                    </div>
                  </td>
                </tr>
                {group.rows.map((row) => (
                  <tr key={row.id}>
                    <td><span className="luxe-value-code">{row.provider_service_id}</span></td>
                    <td>
                       <span className="table-main">
                          <strong>{row.name}</strong>
                          <span className="table-subtext d-block">{row.brand || group.name}</span>
                       </span>
                    </td>
                    <td>
                       <span className="status-badge status-badge-sm info text-capitalize">
                          <i className={`bi bi-${row.platform || 'globe'} me-1`}></i> {row.platform || '-'}
                       </span>
                    </td>
                    <td className="table-number">
                       <div className="table-main">
                          <strong>{row.min_quantity?.toLocaleString('id-ID') || '1'}</strong>
                          <span className="table-subtext">to {row.max_quantity?.toLocaleString('id-ID') || '∞'}</span>
                       </div>
                    </td>
                    <td className="table-number">
                       <span className="luxe-value-strong text-primary">{formatMoney(row.price)}</span>
                    </td>
                    <td className="text-end">
                       <span className={`status-badge ${row.is_active ? 'success' : 'danger'}`}>
                          {row.is_active ? 'Available' : 'Disabled'}
                       </span>
                    </td>
                  </tr>
                ))}
              </Fragment>
            ))}
            {!rows.length && (
              <tr>
                <td colSpan="6" className="empty-cell py-5">
                   <div className="text-center">
                      <i className="bi bi-collection text-muted display-4"></i>
                      <p className="mt-3 text-muted">{loading ? 'Fetching catalog...' : 'No services found.'}</p>
                   </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </section>
      <PaginationBar pagination={pagination} loading={loading} onPageChange={setPage} />
      
      <ConfirmationModal
        show={confirmSync}
        title="Sync Services"
        message={`This will update the ${platformInput || 'global'} catalog from the provider. Service pricing and availability may change.`}
        confirmLabel="Start Sync"
        confirmClassName="btn-primary"
        loading={confirmLoading}
        onCancel={() => setConfirmSync(false)}
        onConfirm={sync}
      />
    </div>
  )
}

export default SMMServices
