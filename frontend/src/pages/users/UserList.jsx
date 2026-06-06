import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import userService from '../../services/userService'
import { getErrorMessage, getListPayload } from '../../services/api'

const UserList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')

  const load = useCallback(async () => {
    const response = await userService.getAll({ search, limit: 50 })
    setRows(getListPayload(response).rows)
  }, [search])

  useEffect(() => {
    load().catch((error) => toast.error(getErrorMessage(error, 'Failed to load users')))
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
        <div>
          <h1>Users</h1>
          <p>Registered accounts and roles.</p>
        </div>
        {hasPermission('users', 'create') && <Link to="/users/new" className="btn btn-primary"><i className="bi bi-plus-lg me-2"></i>New user</Link>}
      </div>

      <form className="filter-pill compact" onSubmit={submitSearch}>
        <i className="bi bi-search"></i>
        <input value={searchInput} onChange={(event) => setSearchInput(event.target.value)} placeholder="Search user" />
        <button className="btn btn-dark" type="submit">Search</button>
        <button className="btn btn-outline-dark" type="button" onClick={resetSearch}>
          <i className="bi bi-x-lg me-2"></i>Reset
        </button>
      </form>

      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr>
              <th>Name</th>
              <th>Email</th>
              <th>Role</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><span className="table-main"><strong>{row.name || '-'}</strong></span></td>
                <td><span className="table-subtext">{row.email}</span></td>
                <td className="table-nowrap">{row.role}</td>
                <td><span className="badge rounded-pill text-bg-light">{row.status || (row.is_active === false ? 'inactive' : 'active')}</span></td>
                <td className="text-end">
                  <span className="table-actions">{hasPermission('users', 'view') && <Link className="btn btn-sm btn-outline-dark" to={`/users/${row.id}`}>Detail</Link>}</span>
                </td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="5" className="empty-cell">No users found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default UserList
