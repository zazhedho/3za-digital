import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import menuService from '../../services/menuService'
import { getErrorMessage, getListPayload } from '../../services/api'

const MenuList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])

  useEffect(() => {
    menuService.getAll({ limit: 100 })
      .then((response) => setRows(getListPayload(response).rows))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load menus')))
  }, [])

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Menus</h1><p>Backend-driven navigation entries.</p></div>
      </div>
      <section className="table-panel">
        <table className="table app-table align-middle">
          <thead><tr><th>Name</th><th>Path</th><th>Icon</th><th>Order</th><th>Active</th><th></th></tr></thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><span className="table-main"><strong>{row.display_name}</strong></span></td>
                <td><span className="table-subtext">{row.path}</span></td>
                <td className="table-nowrap"><i className={`bi ${row.icon || 'bi-circle'} me-2`}></i>{row.icon || '-'}</td>
                <td className="table-number">{row.order_index}</td>
                <td><span className={`badge ${row.is_active ? 'text-bg-success' : 'text-bg-secondary'}`}>{row.is_active ? 'Active' : 'Inactive'}</span></td>
                <td className="text-end"><span className="table-actions">{hasPermission('menus', 'view') && <Link className="btn btn-sm btn-outline-dark" to={`/menus/${row.id}`}>Detail</Link>}</span></td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="6" className="empty-cell">No menus found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default MenuList
