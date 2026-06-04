import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import menuService from '../../services/menuService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const MenuDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [menu, setMenu] = useState(null)

  useEffect(() => {
    menuService.getById(id)
      .then((response) => setMenu(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load menu')))
  }, [id])

  return (
    <div>
      <div className="page-toolbar">
        <div><h1>Menu Detail</h1><p>{menu?.display_name || id}</p></div>
        <div className="toolbar-actions">
          <BackButton fallback="/menus" />
          {hasPermission('menus', 'update') && <Link to={`/menus/${id}/edit`} className="btn btn-primary">Edit</Link>}
        </div>
      </div>
      <section className="panel">
        <div className="detail-grid">
          <span>Name</span><strong>{menu?.name || '-'}</strong>
          <span>Display</span><strong>{menu?.display_name || '-'}</strong>
          <span>Path</span><strong>{menu?.path || '-'}</strong>
          <span>Icon</span><strong>{menu?.icon || '-'}</strong>
          <span>Order</span><strong>{menu?.order_index ?? '-'}</strong>
          <span>Active</span><strong>{menu?.is_active ? 'Yes' : 'No'}</strong>
        </div>
      </section>
    </div>
  )
}

export default MenuDetail
