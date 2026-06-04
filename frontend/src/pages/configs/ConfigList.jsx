import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import appConfigService from '../../services/appConfigService'
import { getErrorMessage, getListPayload } from '../../services/api'

const ConfigList = () => {
  const { hasPermission } = useAuth()
  const [rows, setRows] = useState([])

  useEffect(() => {
    appConfigService.getAll({ limit: 100 })
      .then((response) => setRows(getListPayload(response).rows))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load configs')))
  }, [])

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Configs</h1><p>Application settings.</p></div>
      </div>
      <section className="table-panel">
        <table className="table align-middle">
          <thead><tr><th>Key</th><th>Display</th><th>Category</th><th>Active</th><th></th></tr></thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>{row.config_key}</td>
                <td>{row.display_name}</td>
                <td>{row.category}</td>
                <td>{row.is_active ? 'Yes' : 'No'}</td>
                <td className="text-end">{hasPermission('configs', 'view') && <Link className="btn btn-sm btn-outline-dark" to={`/configs/${row.id}`}>Detail</Link>}</td>
              </tr>
            ))}
            {!rows.length && <tr><td colSpan="5" className="text-center text-muted py-5">No configs found</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  )
}

export default ConfigList
