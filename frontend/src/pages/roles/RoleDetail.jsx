import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { useAuth } from '../../contexts/AuthContext'
import roleService from '../../services/roleService'
import { getErrorMessage } from '../../services/api'
import BackButton from '../../components/common/BackButton'

const RoleDetail = () => {
  const { id } = useParams()
  const { hasPermission } = useAuth()
  const [role, setRole] = useState(null)

  useEffect(() => {
    roleService.getById(id)
      .then((response) => setRole(response.data.data))
      .catch((error) => toast.error(getErrorMessage(error, 'Failed to load role')))
  }, [id])

  return (
    <div>
      <div className="page-toolbar">
        <div>
          <h1>Role Detail</h1>
          <p>{role?.display_name || id}</p>
        </div>
        <div className="toolbar-actions">
          <BackButton fallback="/roles" />
          {hasPermission('roles', 'update') && <Link to={`/roles/${id}/edit`} className="btn btn-primary">Edit</Link>}
        </div>
      </div>
      <section className="panel">
        <div className="detail-grid detail-grid-compact">
          <span>Name</span><strong>{role?.name || '-'}</strong>
          <span>Display</span><strong>{role?.display_name || '-'}</strong>
          <span>Description</span><strong>{role?.description || '-'}</strong>
          <span>System</span><strong>{role?.is_system ? 'Yes' : 'No'}</strong>
          <span>Permissions</span><strong>{role?.permission_ids?.length || 0}</strong>
          <span>Menus</span><strong>{role?.menu_ids?.length || 0}</strong>
        </div>
      </section>
    </div>
  )
}

export default RoleDetail
