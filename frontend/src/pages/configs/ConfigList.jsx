import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import appConfigService from '../../services/appConfigService'
import { getErrorMessage, getListPayload } from '../../services/api'

const ConfigList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')

  const load = useCallback(async () => {
    const response = await appConfigService.getAll({ search, limit: 100 })
    setRows(getListPayload(response).rows)
  }, [search])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load configs')))
  }, [load])

  const submitSearch = (event) => {
    event.preventDefault()
    setSearch(searchInput.trim())
  }

  const resetSearch = () => {
    setSearchInput('')
    setSearch('')
  }

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Configs</h1><p>Application settings.</p></div>
      </div>
      <form className="filter-pill compact" onSubmit={submitSearch}>
        <i className="bi bi-search"></i>
        <input value={searchInput} onChange={(event) => setSearchInput(event.target.value)} placeholder="Search config" />
        <button className="btn btn-dark" type="submit">Search</button>
        <button className="btn btn-outline-dark" type="button" onClick={resetSearch}>
          <i className="bi bi-x-lg me-2"></i>Reset
        </button>
      </form>
      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead><tr><th>Key</th><th>Display</th><th>Category</th><th>Active</th><th></th></tr></thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><span className="table-id">{row.config_key}</span></td>
                <td><span className="table-main"><strong>{row.display_name}</strong></span></td>
                <td className="table-nowrap">{row.category}</td>
                <td><span className={`badge ${row.is_active ? 'text-bg-success' : 'text-bg-secondary'}`}>{row.is_active ? 'Active' : 'Inactive'}</span></td>
                <td className="text-end"><span className="table-actions">{hasPermission('configs', 'view') && <Link className="btn btn-sm btn-outline-dark" to={`/configs/${row.id}`}>Detail</Link>}</span></td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="5" className="empty-cell">No configs found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default ConfigList
