import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import roleService from '../../services/roleService'
import { getErrorMessage, getListPayload } from '../../services/api'

const RoleList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])

  useEffect(() => {
    roleService.getAll({ limit: 50 })
      .then((response) => setRows(getListPayload(response).rows))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load roles')))
  }, [])

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Roles</h1>
          <p>RBAC role registry.</p>
        </div>
        {hasPermission('roles', 'create') && <Link to="/roles/new" className="btn btn-primary"><i className="bi bi-plus-lg me-2"></i>New role</Link>}
      </div>
      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead>
            <tr><th>Name</th><th>Display</th><th>Description</th><th>System</th><th></th></tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><span className="table-main"><strong>{row.name}</strong></span></td>
                <td>{row.display_name || '-'}</td>
                <td><span className="table-subtext">{row.description || '-'}</span></td>
                <td><span className={`badge ${row.is_system ? 'text-bg-secondary' : 'text-bg-light'}`}>{row.is_system ? 'System' : 'Custom'}</span></td>
                <td className="text-end"><span className="table-actions">{hasPermission('roles', 'view') && <Link className="btn btn-sm btn-outline-dark" to={`/roles/${row.id}`}>Detail</Link>}</span></td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="5" className="empty-cell">No roles found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default RoleList
