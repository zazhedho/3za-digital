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
        <table className="table align-middle">
          <thead>
            <tr><th>Name</th><th>Display</th><th>Description</th><th>System</th><th></th></tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>{row.name}</td>
                <td>{row.display_name || '-'}</td>
                <td>{row.description || '-'}</td>
                <td>{row.is_system ? 'Yes' : 'No'}</td>
                <td className="text-end">{hasPermission('roles', 'view') && <Link className="btn btn-sm btn-outline-dark" to={`/roles/${row.id}`}>Detail</Link>}</td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="5" className="text-center text-muted py-5">No roles found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default RoleList
