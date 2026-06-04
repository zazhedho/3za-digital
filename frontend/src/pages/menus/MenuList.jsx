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
        <table className="table align-middle">
          <thead><tr><th>Name</th><th>Path</th><th>Icon</th><th>Order</th><th>Active</th><th></th></tr></thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>{row.display_name}</td>
                <td>{row.path}</td>
                <td><i className={`bi ${row.icon || 'bi-circle'}`}></i> {row.icon}</td>
                <td>{row.order_index}</td>
                <td>{row.is_active ? 'Yes' : 'No'}</td>
                <td className="text-end">{hasPermission('menus', 'view') && <Link className="btn btn-sm btn-outline-dark" to={`/menus/${row.id}`}>Detail</Link>}</td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="6" className="text-center text-muted py-5">No menus found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default MenuList
